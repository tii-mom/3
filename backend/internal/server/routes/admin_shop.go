package routes

import (
	"crypto/rand"
	"encoding/hex"
	"image"
	_ "image/gif"  // 注册 GIF 解码器，供 image.DecodeConfig 读取尺寸
	_ "image/jpeg" // 注册 JPEG 解码器
	_ "image/png"  // 注册 PNG 解码器
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const (
	// 上传上限直接对齐「首页展示所需的图片大小」，而不是沿用一个与展示位无关的通用值：
	//   - 商品图在首页只是 46px 方块（PlanCard → BrandLogo），前端会先裁成 512×512 再上传；
	//   - 轮播图铺满卡片宽度（约 540px），允许稍大的原图。
	// 前端压缩后通常 <200KB，这里的上限是给直连接口/脚本上传兜底。
	maxShopAssetProductUploadSize = 1 << 20 // 1MB
	maxShopAssetBannerUploadSize  = 3 << 20 // 3MB
	// 像素边长上限：挡住「体积很小、解码后巨大」的压缩炸弹
	maxShopAssetPixelDim = 4096
)

// shopAssetPurpose 归一化素材用途；缺省按商品图处理（限制更严，安全侧默认）。
func shopAssetPurpose(c *gin.Context) string {
	if strings.EqualFold(strings.TrimSpace(c.PostForm("purpose")), "banner") {
		return "banner"
	}
	return "product"
}

func shopAssetUploadLimit(purpose string) int64 {
	if purpose == "banner" {
		return maxShopAssetBannerUploadSize
	}
	return maxShopAssetProductUploadSize
}

func shopAssetUploadLimitMessage(purpose string) string {
	if purpose == "banner" {
		return "轮播图不能超过 3MB"
	}
	return "商品图不能超过 1MB"
}

func RegisterAdminShopRoutes(
	v1 *gin.RouterGroup,
	shopHandler *admin.ShopHandler,
	adminAuth middleware.AdminAuthMiddleware,
	auditLog middleware.AuditLogMiddleware,
	settingService *service.SettingService,
	dataDir string,
) {
	v1.GET("/shop/assets/:filename", serveShopAsset(dataDir))

	adminGroup := v1.Group("/admin/shop")
	adminGroup.Use(gin.HandlerFunc(adminAuth))
	adminGroup.Use(gin.HandlerFunc(auditLog))
	adminGroup.Use(middleware.AdminComplianceGuard(settingService))
	{
		adminGroup.GET("/categories", shopHandler.ListCategories)
		adminGroup.POST("/categories", shopHandler.CreateCategory)
		adminGroup.PUT("/categories/:id", shopHandler.UpdateCategory)
		adminGroup.DELETE("/categories/:id", shopHandler.DeleteCategory)

		adminGroup.GET("/products", shopHandler.ListProducts)
		adminGroup.POST("/products", shopHandler.CreateProduct)
		adminGroup.PUT("/products/:id", shopHandler.UpdateProduct)
		adminGroup.DELETE("/products/:id", shopHandler.DeleteProduct)

		adminGroup.GET("/banners", shopHandler.ListBanners)
		adminGroup.POST("/banners", shopHandler.CreateBanner)
		adminGroup.PUT("/banners/:id", shopHandler.UpdateBanner)
		adminGroup.DELETE("/banners/:id", shopHandler.DeleteBanner)

		adminGroup.GET("/orders", shopHandler.ListOrders)
		adminGroup.POST("/orders/:id/fulfill", shopHandler.FulfillOrder)

		adminGroup.POST("/assets", uploadShopAsset(dataDir))
	}
}

func uploadShopAsset(dataDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		purpose := shopAssetPurpose(c)
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_FILE", "message": "请选择要上传的图片"})
			return
		}
		if file.Size <= 0 || file.Size > shopAssetUploadLimit(purpose) {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_FILE_SIZE", "message": shopAssetUploadLimitMessage(purpose)})
			return
		}

		ext := strings.ToLower(filepath.Ext(file.Filename))
		contentType, err := detectShopAssetContentType(file)
		if err != nil || !isAllowedShopAsset(ext, contentType) {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_FILE_TYPE", "message": "仅支持 JPG、PNG、WebP、GIF 图片"})
			return
		}

		// 尺寸校验：WebP 走自定义头解析（标准库不解码 WebP），其余交给 image.DecodeConfig。
		// 解不出尺寸时不拦截——体积上限已经兜住了绝大多数异常，不该因为格式太新而误伤。
		if width, height, ok := detectShopAssetDimensions(file); ok {
			if width > maxShopAssetPixelDim || height > maxShopAssetPixelDim {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    "INVALID_FILE_DIMENSIONS",
					"message": "图片尺寸过大，请压缩到 4096px 以内后再上传",
				})
				return
			}
		}

		dir := filepath.Join(dataDir, "shop-assets")
		if err := os.MkdirAll(dir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": "UPLOAD_FAILED", "message": "创建上传目录失败"})
			return
		}

		name := time.Now().UTC().Format("20060102-150405") + "-" + randomHex(8) + ext
		dst := filepath.Join(dir, name)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": "UPLOAD_FAILED", "message": "图片保存失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{"url": "/api/v1/shop/assets/" + name},
		})
	}
}

func serveShopAsset(dataDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := filepath.Base(c.Param("filename"))
		if filename == "." || filename == "/" || strings.Contains(filename, "..") {
			c.Status(http.StatusNotFound)
			return
		}
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Cache-Control", "public, max-age=604800, immutable")
		path := filepath.Join(dataDir, "shop-assets", filename)
		c.File(path)
	}
}

func detectShopAssetContentType(fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.ToLower(http.DetectContentType(buf[:n])), nil
}

func isAllowedShopAsset(ext, contentType string) bool {
	switch ext {
	case ".jpg", ".jpeg":
		return strings.HasPrefix(contentType, "image/jpeg")
	case ".png":
		return strings.HasPrefix(contentType, "image/png")
	case ".webp":
		return strings.HasPrefix(contentType, "image/webp")
	case ".gif":
		return strings.HasPrefix(contentType, "image/gif")
	default:
		return false
	}
}

// detectShopAssetDimensions 读取图片像素尺寸。
// 第三个返回值为 false 表示「无法判定」，调用方应当放行而不是拒收。
func detectShopAssetDimensions(fileHeader *multipart.FileHeader) (int, int, bool) {
	file, err := fileHeader.Open()
	if err != nil {
		return 0, 0, false
	}
	defer func() { _ = file.Close() }()

	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return 0, 0, false
	}
	header = header[:n]

	// WebP 不在标准库解码器内：自己读容器头，避免前端压缩产出的 WebP 绕过尺寸校验
	if width, height, ok := webpDimensions(header); ok {
		return width, height, true
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return 0, 0, false
	}
	cfg, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, false
	}
	return cfg.Width, cfg.Height, true
}

// webpDimensions 从 RIFF/WEBP 容器头解析画布尺寸，支持 VP8X / VP8 / VP8L 三种块。
// 只依赖容器头，不需要解码像素。无法识别时返回 ok=false。
func webpDimensions(data []byte) (int, int, bool) {
	if len(data) < 16 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WEBP" {
		return 0, 0, false
	}
	switch string(data[12:16]) {
	case "VP8X":
		// 24 位小端存放「宽-1 / 高-1」
		if len(data) < 30 {
			return 0, 0, false
		}
		width := 1 + int(data[24]) + int(data[25])<<8 + int(data[26])<<16
		height := 1 + int(data[27]) + int(data[28])<<8 + int(data[29])<<16
		return width, height, true
	case "VP8 ":
		// 关键帧起始码 0x9d 0x01 0x2a 后的两个 16 位小端字段
		if len(data) < 30 || data[23] != 0x9d || data[24] != 0x01 || data[25] != 0x2a {
			return 0, 0, false
		}
		width := int(data[26]) | int(data[27])<<8
		height := int(data[28]) | int(data[29])<<8
		return width & 0x3fff, height & 0x3fff, true
	case "VP8L":
		if len(data) < 25 || data[20] != 0x2f {
			return 0, 0, false
		}
		b0, b1, b2, b3 := int(data[21]), int(data[22]), int(data[23]), int(data[24])
		width := 1 + (((b1 & 0x3f) << 8) | b0)
		height := 1 + (((b3 & 0x0f) << 10) | (b2 << 2) | ((b1 & 0xc0) >> 6))
		return width, height, true
	}
	return 0, 0, false
}

func randomHex(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "fallback"
	}
	return hex.EncodeToString(buf)
}
