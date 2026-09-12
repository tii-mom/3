package routes

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/require"
)

// ---------- WebP 容器头构造（仅用于测试尺寸解析） ----------

func webpRIFF(chunk string, size int) []byte {
	buf := make([]byte, 0, 32)
	buf = append(buf, []byte("RIFF")...)
	buf = append(buf, make([]byte, 4)...)
	buf = append(buf, []byte("WEBP")...)
	buf = append(buf, []byte(chunk)...)
	sizeBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBuf, uint32(size))
	buf = append(buf, sizeBuf...)
	return buf
}

func webpVP8X(width, height uint32) []byte {
	buf := webpRIFF("VP8X", 10)
	buf = append(buf, 0, 0, 0, 0) // flags + reserved
	w, h := width-1, height-1
	buf = append(buf, byte(w), byte(w>>8), byte(w>>16))
	buf = append(buf, byte(h), byte(h>>8), byte(h>>16))
	return buf
}

func webpLossy(width, height uint16) []byte {
	buf := webpRIFF("VP8 ", 10)
	buf = append(buf, 0, 0, 0)          // frame tag
	buf = append(buf, 0x9d, 0x01, 0x2a) // 关键帧起始码
	dim := make([]byte, 4)
	binary.LittleEndian.PutUint16(dim[0:2], width)
	binary.LittleEndian.PutUint16(dim[2:4], height)
	buf = append(buf, dim...)
	return buf
}

func webpLossless(width, height uint32) []byte {
	buf := webpRIFF("VP8L", 5)
	buf = append(buf, 0x2f)
	bits := (width - 1) | ((height - 1) << 14)
	packed := make([]byte, 4)
	binary.LittleEndian.PutUint32(packed, bits)
	buf = append(buf, packed...)
	return buf
}

func TestWebpDimensionsReadsEveryChunkType(t *testing.T) {
	cases := []struct {
		name          string
		data          []byte
		width, height int
		ok            bool
	}{
		{name: "vp8x", data: webpVP8X(3840, 2160), width: 3840, height: 2160, ok: true},
		{name: "lossy", data: webpLossy(512, 512), width: 512, height: 512, ok: true},
		{name: "lossless", data: webpLossless(1920, 1080), width: 1920, height: 1080, ok: true},
		{name: "not webp", data: []byte("RIFF\x00\x00\x00\x00WAVEfmt "), ok: false},
		{name: "too short", data: []byte("RIFF"), ok: false},
		{name: "unknown chunk", data: append(webpRIFF("ABCD", 4), make([]byte, 12)...), ok: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			width, height, ok := webpDimensions(tc.data)
			require.Equal(t, tc.ok, ok)
			if tc.ok {
				require.Equal(t, tc.width, width)
				require.Equal(t, tc.height, height)
			}
		})
	}
}

func TestWebpLossyRejectsBrokenStartCode(t *testing.T) {
	data := webpLossy(512, 512)
	data[23] = 0x00
	_, _, ok := webpDimensions(data)
	require.False(t, ok)
}

// ---------- 尺寸探测（真实 multipart 文件头） ----------

func multipartFileHeader(t *testing.T, filename, contentType string, data []byte) *multipart.FileHeader {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	require.NoError(t, err)
	_, err = part.Write(data)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	require.NoError(t, req.ParseMultipartForm(16<<20))
	_, fileHeader, err := req.FormFile("file")
	require.NoError(t, err)
	return fileHeader
}

func encodePNG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	buf := &bytes.Buffer{}
	require.NoError(t, png.Encode(buf, img))
	return buf.Bytes()
}

func TestDetectShopAssetDimensions(t *testing.T) {
	t.Run("png 通过标准库解码", func(t *testing.T) {
		header := multipartFileHeader(t, "big.png", "image/png", encodePNG(t, 5000, 12))
		width, height, ok := detectShopAssetDimensions(header)
		require.True(t, ok)
		require.Equal(t, 5000, width)
		require.Equal(t, 12, height)
	})

	t.Run("webp 走后缀头解析", func(t *testing.T) {
		header := multipartFileHeader(t, "shot.webp", "image/webp", webpVP8X(768, 512))
		width, height, ok := detectShopAssetDimensions(header)
		require.True(t, ok)
		require.Equal(t, 768, width)
		require.Equal(t, 512, height)
	})

	t.Run("无法识别时放行而不是拒绝", func(t *testing.T) {
		header := multipartFileHeader(t, "broken.png", "image/png", []byte("not an image at all"))
		_, _, ok := detectShopAssetDimensions(header)
		require.False(t, ok)
	})
}

// ---------- 用途与上限 ----------

func TestShopAssetUploadLimitsByPurpose(t *testing.T) {
	require.Equal(t, int64(maxShopAssetProductUploadSize), shopAssetUploadLimit("product"))
	require.Equal(t, int64(maxShopAssetBannerUploadSize), shopAssetUploadLimit("banner"))
	// 未知用途回落到更严格的商品图档
	require.Equal(t, int64(maxShopAssetProductUploadSize), shopAssetUploadLimit(""))
	require.Contains(t, shopAssetUploadLimitMessage("banner"), "轮播图")
	require.Contains(t, shopAssetUploadLimitMessage("product"), "商品图")
}

func TestIsAllowedShopAsset(t *testing.T) {
	require.True(t, isAllowedShopAsset(".webp", "image/webp"))
	require.True(t, isAllowedShopAsset(".jpeg", "image/jpeg"))
	require.False(t, isAllowedShopAsset(".webp", "image/png"))
	require.False(t, isAllowedShopAsset(".svg", "image/svg+xml"))
}
