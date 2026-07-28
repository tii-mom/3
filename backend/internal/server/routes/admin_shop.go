package routes

import (
	"crypto/rand"
	"encoding/hex"
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

const maxShopAssetUploadSize = 5 << 20 // 5MB

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
		adminGroup.GET("/products", shopHandler.ListProducts)
		adminGroup.POST("/products", shopHandler.CreateProduct)
		adminGroup.PUT("/products/:id", shopHandler.UpdateProduct)
		adminGroup.DELETE("/products/:id", shopHandler.DeleteProduct)

		adminGroup.GET("/banners", shopHandler.ListBanners)
		adminGroup.POST("/banners", shopHandler.CreateBanner)
		adminGroup.PUT("/banners/:id", shopHandler.UpdateBanner)
		adminGroup.DELETE("/banners/:id", shopHandler.DeleteBanner)

		adminGroup.GET("/orders", shopHandler.ListOrders)

		adminGroup.POST("/assets", uploadShopAsset(dataDir))
	}
}

func uploadShopAsset(dataDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_FILE", "message": "请选择要上传的图片"})
			return
		}
		if file.Size <= 0 || file.Size > maxShopAssetUploadSize {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_FILE_SIZE", "message": "图片大小不能超过 5MB"})
			return
		}

		ext := strings.ToLower(filepath.Ext(file.Filename))
		contentType, err := detectShopAssetContentType(file)
		if err != nil || !isAllowedShopAsset(ext, contentType) {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_FILE_TYPE", "message": "仅支持 JPG、PNG、WebP、GIF 图片"})
			return
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
	defer file.Close()
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

func randomHex(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "fallback"
	}
	return hex.EncodeToString(buf)
}
