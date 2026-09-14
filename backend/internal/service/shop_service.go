package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/Wei-Shaw/sub2api/internal/pkg/creditctx"
	"github.com/Wei-Shaw/sub2api/internal/pkg/creditledger"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/shopspring/decimal"
)

const (
	ShopProductTypeVirtual            = "virtual"
	ShopProductTypePlatformUSDBalance = "platform_usd_balance"
)

// 商品交付模式（迁移 206）
const (
	ShopFulfillmentManual          = "manual"           // 人工处理，仅需备注
	ShopFulfillmentSessionTopup    = "session_topup"    // 代充值：用户提交 ChatGPT Session
	ShopFulfillmentAccountDelivery = "account_delivery" // 成品号：后台人工发账号
	ShopFulfillmentRental          = "rental"           // 租号：按时长租赁，后台人工发账号
)

// 商品品类种子 slug（迁移 207 引入，迁移 208 起集合改由 shop_categories 表定义）。
// 后台可自行新增品类，因此这里只保留「兜底 slug」与文档意义，不再作为校验白名单。
const (
	ShopCategoryGPTTopup   = "gpt_topup"   // GPT / ChatGPT 代充值
	ShopCategoryGPTAccount = "gpt_account" // GPT / ChatGPT 成品号
	ShopCategoryXPremium   = "x_premium"   // X（Twitter）蓝 V / Premium
	ShopCategoryGemini     = "gemini"      // Gemini 会员
	ShopCategoryCodex      = "codex"       // Codex 相关
	ShopCategoryOther      = "other"       // 其他（未分类商品的兜底值）
)

// shopCategorySlugPattern 品类标识格式：小写字母开头，仅含小写字母/数字/下划线，最长 32。
// slug 会进入前端 DOM id（cat-<slug>）与锚点 URL，必须限制字符集。
// 品类是否「存在」由 shop_categories 表判定（迁移 208 起），不再由代码枚举。
var shopCategorySlugPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)

func isValidCategorySlug(slug string) bool {
	return shopCategorySlugPattern.MatchString(slug)
}

// shopProductColumnList 统一商品 SELECT 列，避免各处 SQL 与扫描顺序不一致。
// 顺序必须与 scanShopProduct 的 Scan 顺序严格一致。
var shopProductColumnList = []string{
	"id", "name", "description", "image_url", "product_type", "price_cny_minor",
	"original_price_cny_minor", "grant_usd_amount::text", "stock_quantity", "sold_count",
	"commission_bps", "status", "sort_order", "fulfillment_mode", "delivery_form_hint",
	"badge_text", "spec_label", "highlight", "category", "gallery_json",
	"created_at", "updated_at",
}

// shopProductColumnsWith 生成带表别名的列清单（公开列表要 JOIN 品类表，必须消歧义）。
// alias 为空时等价于旧的无别名写法。
func shopProductColumnsWith(alias string) string {
	parts := make([]string, 0, len(shopProductColumnList))
	for _, col := range shopProductColumnList {
		if alias == "" {
			parts = append(parts, col)
			continue
		}
		parts = append(parts, alias+"."+col)
	}
	return strings.Join(parts, ", ")
}

var shopProductColumns = shopProductColumnsWith("")

// 系统游客账号邮箱（迁移 206 插入），用于承载免登录订单。
const shopGuestUserEmail = "guest@internal.3api.invalid"

func isKnownFulfillmentMode(mode string) bool {
	switch mode {
	case ShopFulfillmentManual, ShopFulfillmentSessionTopup, ShopFulfillmentAccountDelivery, ShopFulfillmentRental:
		return true
	}
	return false
}

type ShopService struct {
	db             *sql.DB
	paymentService *PaymentService
	encryptor      SecretEncryptor
	guestLimiter   *guestOrderLimiter
}

func NewShopService(db *sql.DB) *ShopService {
	return &ShopService{db: db, guestLimiter: newGuestOrderLimiter(db)}
}

func (s *ShopService) SetPaymentService(paymentService *PaymentService) {
	s.paymentService = paymentService
}

// SetSecretEncryptor 注入用于加密存储交付资料（Session 凭证等）的加密器。
// 未注入时公开下单接口会拒绝创建订单，避免敏感信息明文落库。
func (s *ShopService) SetSecretEncryptor(encryptor SecretEncryptor) {
	s.encryptor = encryptor
}

type ShopProduct struct {
	ID                    int64    `json:"id"`
	Name                  string   `json:"name"`
	Description           string   `json:"description"`
	ImageURL              string   `json:"image_url"`
	ProductType           string   `json:"product_type"`
	PriceCNYMinor         int64    `json:"price_cny_minor"`
	OriginalPriceCNYMinor int64    `json:"original_price_cny_minor"`
	GrantUSDAmount        string   `json:"grant_usd_amount"`
	StockQuantity         *int64   `json:"stock_quantity,omitempty"`
	SoldCount             int64    `json:"sold_count"`
	CommissionBPS         int      `json:"commission_bps"`
	Status                string   `json:"status"`
	SortOrder             int      `json:"sort_order"`
	FulfillmentMode       string   `json:"fulfillment_mode"`
	DeliveryFormHint      string   `json:"delivery_form_hint"`
	BadgeText             string   `json:"badge_text"`
	SpecLabel             string   `json:"spec_label"`
	Highlight             bool     `json:"highlight"`
	Category              string   `json:"category"`
	Gallery               []string `json:"gallery"`
	CreatedAt             string   `json:"created_at,omitempty"`
	UpdatedAt             string   `json:"updated_at,omitempty"`
}

type ShopBanner struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle"`
	ImageURL   string `json:"image_url"`
	ButtonText string `json:"button_text"`
	ProductID  *int64 `json:"product_id,omitempty"`
	Enabled    bool   `json:"enabled"`
	SortOrder  int    `json:"sort_order"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

type ShopOrder struct {
	ID                    int64  `json:"id"`
	UserID                int64  `json:"user_id"`
	ProductID             int64  `json:"product_id"`
	PaymentOrderID        *int64 `json:"payment_order_id,omitempty"`
	Status                string `json:"status"`
	FulfillmentStatus     string `json:"fulfillment_status"`
	CommissionStatus      string `json:"commission_status"`
	SnapshotName          string `json:"snapshot_name"`
	SnapshotDescription   string `json:"snapshot_description"`
	SnapshotImageURL      string `json:"snapshot_image_url"`
	SnapshotProductType   string `json:"snapshot_product_type"`
	SnapshotPriceCNYMinor int64  `json:"snapshot_price_cny_minor"`
	// WalletAppliedCNYMinor = 本单用人民币返点余额抵扣的金额；
	// PayableCNYMinor = 还需外部支付（微信/支付宝）的金额。
	// 恒等式：PayableCNYMinor = SnapshotPriceCNYMinor - WalletAppliedCNYMinor。
	WalletAppliedCNYMinor   int64   `json:"wallet_applied_cny_minor"`
	PayableCNYMinor         int64   `json:"payable_cny_minor"`
	SnapshotGrantUSD        string  `json:"snapshot_grant_usd_amount"`
	SnapshotCommissionBPS   int     `json:"snapshot_commission_bps"`
	FulfillmentNote         string  `json:"fulfillment_note"`
	UserEmail               string  `json:"user_email,omitempty"`
	OrderNo                 string  `json:"order_no,omitempty"`
	GuestToken              string  `json:"guest_token,omitempty"`
	GuestContact            string  `json:"guest_contact,omitempty"`
	SnapshotFulfillmentMode string  `json:"snapshot_fulfillment_mode,omitempty"`
	DeliveryHint            string  `json:"delivery_hint,omitempty"`
	DeliverySubmittedAt     *string `json:"delivery_submitted_at,omitempty"`
	RentalDuration          string  `json:"rental_duration,omitempty"`
	CreatedAt               string  `json:"created_at"`
	PaidAt                  *string `json:"paid_at,omitempty"`
	FulfilledAt             *string `json:"fulfilled_at,omitempty"`
}

type CreateShopOrderResult struct {
	ShopOrderID int64                `json:"shop_order_id"`
	Payment     *CreateOrderResponse `json:"payment"`
	// WalletAppliedCNYMinor 本单抵扣的返点余额；PayableCNYMinor 还需外部支付的金额。
	WalletAppliedCNYMinor int64 `json:"wallet_applied_cny_minor"`
	PayableCNYMinor       int64 `json:"payable_cny_minor"`
	// FullyPaidByWallet 为 true 时没有外部支付环节：订单已经直接进入交付，
	// 前端不要再去拉起支付渠道，直接跳订单页。
	FullyPaidByWallet bool `json:"fully_paid_by_wallet"`
}

// ShopWalletBalance 是「我的返点余额」快照，用于下单页展示可用抵扣。
type ShopWalletBalance struct {
	Enabled           bool  `json:"enabled"`
	AvailableCNYMinor int64 `json:"available_cny_minor"`
	FrozenCNYMinor    int64 `json:"frozen_cny_minor"`
}

type UpsertShopProductInput struct {
	Name                  string
	Description           string
	ImageURL              string
	ProductType           string
	PriceCNYMinor         int64
	OriginalPriceCNYMinor int64
	GrantUSDAmount        string
	StockQuantity         *int64
	CommissionBPS         int
	Status                string
	SortOrder             int
	FulfillmentMode       string
	DeliveryFormHint      string
	BadgeText             string
	SpecLabel             string
	Highlight             bool
	Category              string
	Gallery               []string
}

type UpsertShopBannerInput struct {
	Title      string
	Subtitle   string
	ImageURL   string
	ButtonText string
	ProductID  *int64
	Enabled    bool
	SortOrder  int
}

func (s *ShopService) ListPublicProducts(ctx context.Context) ([]ShopProduct, error) {
	return s.listProducts(ctx, false)
}

func (s *ShopService) AdminListProducts(ctx context.Context) ([]ShopProduct, error) {
	return s.listProducts(ctx, true)
}

func (s *ShopService) listProducts(ctx context.Context, admin bool) ([]ShopProduct, error) {
	where := `p.tenant_id = 1 AND p.deleted_at IS NULL`
	if !admin {
		// 公开列表只出「已发布」且品类未被后台停用的商品。
		// LEFT JOIN + COALESCE：品类已被删除或 slug 不存在时仍视为可见，避免脏数据让商品凭空消失。
		where += ` AND p.status = 'published' AND COALESCE(c.enabled, TRUE)`
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT `+shopProductColumnsWith("p")+`
FROM shop_products p
LEFT JOIN shop_categories c ON c.tenant_id = p.tenant_id AND c.slug = p.category
WHERE `+where+`
ORDER BY p.sort_order ASC, p.id DESC`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []ShopProduct{}
	for rows.Next() {
		item, err := scanShopProduct(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *ShopService) ListPublicBanners(ctx context.Context) ([]ShopBanner, error) {
	return s.listBanners(ctx, false)
}

func (s *ShopService) AdminListBanners(ctx context.Context) ([]ShopBanner, error) {
	return s.listBanners(ctx, true)
}

func (s *ShopService) listBanners(ctx context.Context, admin bool) ([]ShopBanner, error) {
	where := `tenant_id = 1 AND deleted_at IS NULL`
	if !admin {
		where += ` AND enabled = TRUE`
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, title, subtitle, image_url, button_text, product_id, enabled, sort_order, created_at, updated_at
FROM shop_banners
WHERE `+where+`
ORDER BY sort_order ASC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []ShopBanner{}
	for rows.Next() {
		var item ShopBanner
		var productID sql.NullInt64
		var created, updated time.Time
		if err := rows.Scan(&item.ID, &item.Title, &item.Subtitle, &item.ImageURL, &item.ButtonText, &productID, &item.Enabled, &item.SortOrder, &created, &updated); err != nil {
			return nil, err
		}
		if productID.Valid {
			item.ProductID = &productID.Int64
		}
		item.CreatedAt = created.Format(time.RFC3339)
		item.UpdatedAt = updated.Format(time.RFC3339)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *ShopService) CreateProduct(ctx context.Context, in UpsertShopProductInput) (*ShopProduct, error) {
	in = normalizeShopProductInput(in)
	if err := validateShopProductInput(in); err != nil {
		return nil, err
	}
	if err := s.ensureCategoryExists(ctx, in.Category); err != nil {
		return nil, err
	}
	var item ShopProduct
	row := s.db.QueryRowContext(ctx, `
INSERT INTO shop_products (tenant_id, name, description, image_url, product_type, price_cny_minor,
    original_price_cny_minor, grant_usd_amount, stock_quantity, commission_bps, status, sort_order,
    fulfillment_mode, delivery_form_hint, badge_text, spec_label, highlight, category, gallery_json)
VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
RETURNING `+shopProductColumns,
		in.Name, in.Description, in.ImageURL, in.ProductType, in.PriceCNYMinor, in.OriginalPriceCNYMinor,
		normalizeDecimalString(in.GrantUSDAmount), in.StockQuantity, in.CommissionBPS, in.Status, in.SortOrder,
		in.FulfillmentMode, in.DeliveryFormHint, in.BadgeText, in.SpecLabel, in.Highlight,
		in.Category, encodeGalleryJSON(in.Gallery))
	item, err := scanShopProduct(row)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *ShopService) UpdateProduct(ctx context.Context, id int64, in UpsertShopProductInput) (*ShopProduct, error) {
	if id <= 0 {
		return nil, infraerrors.BadRequest("INVALID_ID", "invalid product id")
	}
	in = normalizeShopProductInput(in)
	if err := validateShopProductInput(in); err != nil {
		return nil, err
	}
	if err := s.ensureCategoryExists(ctx, in.Category); err != nil {
		return nil, err
	}
	var item ShopProduct
	row := s.db.QueryRowContext(ctx, `
UPDATE shop_products
SET name = $2, description = $3, image_url = $4, product_type = $5, price_cny_minor = $6,
    original_price_cny_minor = $7, grant_usd_amount = $8, stock_quantity = $9, commission_bps = $10,
    status = $11, sort_order = $12, fulfillment_mode = $13, delivery_form_hint = $14, badge_text = $15,
    spec_label = $16, highlight = $17, category = $18, gallery_json = $19, updated_at = NOW()
WHERE tenant_id = 1 AND id = $1 AND deleted_at IS NULL
RETURNING `+shopProductColumns,
		id, in.Name, in.Description, in.ImageURL, in.ProductType, in.PriceCNYMinor, in.OriginalPriceCNYMinor,
		normalizeDecimalString(in.GrantUSDAmount), in.StockQuantity, in.CommissionBPS, in.Status, in.SortOrder,
		in.FulfillmentMode, in.DeliveryFormHint, in.BadgeText, in.SpecLabel, in.Highlight,
		in.Category, encodeGalleryJSON(in.Gallery))
	item, err := scanShopProduct(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, infraerrors.NotFound("SHOP_PRODUCT_NOT_FOUND", "product not found")
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *ShopService) DeleteProduct(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `UPDATE shop_products SET deleted_at = NOW(), status = 'archived', updated_at = NOW() WHERE tenant_id = 1 AND id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return infraerrors.NotFound("SHOP_PRODUCT_NOT_FOUND", "product not found")
	}
	return nil
}

func (s *ShopService) CreateBanner(ctx context.Context, in UpsertShopBannerInput) (*ShopBanner, error) {
	if strings.TrimSpace(in.Title) == "" {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "banner title is required")
	}
	if strings.TrimSpace(in.ButtonText) == "" {
		in.ButtonText = "立即查看"
	}
	var item ShopBanner
	var productID sql.NullInt64
	var created, updated time.Time
	err := s.db.QueryRowContext(ctx, `
INSERT INTO shop_banners (tenant_id, title, subtitle, image_url, button_text, product_id, enabled, sort_order)
VALUES (1, $1, $2, $3, $4, $5, $6, $7)
RETURNING id, title, subtitle, image_url, button_text, product_id, enabled, sort_order, created_at, updated_at`,
		strings.TrimSpace(in.Title), strings.TrimSpace(in.Subtitle), strings.TrimSpace(in.ImageURL), strings.TrimSpace(in.ButtonText), in.ProductID, in.Enabled, in.SortOrder).
		Scan(&item.ID, &item.Title, &item.Subtitle, &item.ImageURL, &item.ButtonText, &productID, &item.Enabled, &item.SortOrder, &created, &updated)
	if err != nil {
		return nil, err
	}
	if productID.Valid {
		item.ProductID = &productID.Int64
	}
	item.CreatedAt = created.Format(time.RFC3339)
	item.UpdatedAt = updated.Format(time.RFC3339)
	return &item, nil
}

func (s *ShopService) UpdateBanner(ctx context.Context, id int64, in UpsertShopBannerInput) (*ShopBanner, error) {
	if strings.TrimSpace(in.Title) == "" {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "banner title is required")
	}
	if strings.TrimSpace(in.ButtonText) == "" {
		in.ButtonText = "立即查看"
	}
	var item ShopBanner
	var productID sql.NullInt64
	var created, updated time.Time
	err := s.db.QueryRowContext(ctx, `
UPDATE shop_banners
SET title = $2, subtitle = $3, image_url = $4, button_text = $5, product_id = $6, enabled = $7, sort_order = $8, updated_at = NOW()
WHERE tenant_id = 1 AND id = $1 AND deleted_at IS NULL
RETURNING id, title, subtitle, image_url, button_text, product_id, enabled, sort_order, created_at, updated_at`,
		id, strings.TrimSpace(in.Title), strings.TrimSpace(in.Subtitle), strings.TrimSpace(in.ImageURL), strings.TrimSpace(in.ButtonText), in.ProductID, in.Enabled, in.SortOrder).
		Scan(&item.ID, &item.Title, &item.Subtitle, &item.ImageURL, &item.ButtonText, &productID, &item.Enabled, &item.SortOrder, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, infraerrors.NotFound("SHOP_BANNER_NOT_FOUND", "banner not found")
	}
	if err != nil {
		return nil, err
	}
	if productID.Valid {
		item.ProductID = &productID.Int64
	}
	item.CreatedAt = created.Format(time.RFC3339)
	item.UpdatedAt = updated.Format(time.RFC3339)
	return &item, nil
}

func (s *ShopService) DeleteBanner(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `UPDATE shop_banners SET deleted_at = NOW(), enabled = FALSE, updated_at = NOW() WHERE tenant_id = 1 AND id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return infraerrors.NotFound("SHOP_BANNER_NOT_FOUND", "banner not found")
	}
	return nil
}

func (s *ShopService) CreateOrderAndPayment(ctx context.Context, userID int64, productID int64, paymentType, returnURL, clientIP, srcHost, srcURL, locale string, isMobile, isWeChatBrowser bool, openID string, useWallet bool) (*CreateShopOrderResult, error) {
	shopOrderID, payableMinor, appliedMinor, err := s.createPendingShopOrder(ctx, userID, productID, useWallet)
	if err != nil {
		return nil, err
	}
	// 全额抵扣：没有外部支付环节，直接完成交付（下单 → 发货走同一条幂等链路）。
	if payableMinor <= 0 {
		if err := s.FulfillWalletPaidOrder(ctx, shopOrderID); err != nil {
			// 交付失败（如库存被并发抢空）时把订单作废并退回抵扣的余额，
			// 不能出现「余额已扣、货没发」的悬空订单。
			s.abortPendingShopOrder(ctx, shopOrderID, "failed")
			return nil, err
		}
		return &CreateShopOrderResult{
			ShopOrderID:           shopOrderID,
			WalletAppliedCNYMinor: appliedMinor,
			PayableCNYMinor:       0,
			FullyPaidByWallet:     true,
		}, nil
	}
	if s.paymentService == nil {
		// 支付系统不可用：订单不可能再被支付。必须在这里就作废订单并退回已抵扣的
		// 返点余额，否则会留下一张永远 pending、余额却被扣掉的悬空订单
		// （没有外部支付单，也就没有任何超时事件会来关它）。
		s.abortPendingShopOrder(ctx, shopOrderID, "failed")
		return nil, infraerrors.Forbidden("PAYMENT_UNAVAILABLE", "payment system is unavailable")
	}
	// 有外部应付金额时才要求支付方式；全额抵扣的订单没有支付环节。
	if strings.TrimSpace(paymentType) == "" {
		s.abortPendingShopOrder(ctx, shopOrderID, "failed")
		return nil, infraerrors.BadRequest("PAYMENT_TYPE_REQUIRED", "payment type is required")
	}
	paymentResp, err := s.paymentService.CreateOrder(ctx, CreateOrderRequest{
		UserID:          userID,
		Amount:          float64(payableMinor) / 100,
		PaymentType:     paymentType,
		OpenID:          openID,
		ClientIP:        clientIP,
		IsMobile:        isMobile,
		IsWeChatBrowser: isWeChatBrowser,
		SrcHost:         srcHost,
		SrcURL:          srcURL,
		ReturnURL:       returnURL,
		PaymentSource:   "shop",
		OrderType:       payment.OrderTypeShop,
		ShopOrderID:     shopOrderID,
		Locale:          locale,
	})
	if err != nil {
		// 支付单创建失败：订单作废，同时把已经抵扣的返点余额退回。
		s.abortPendingShopOrder(ctx, shopOrderID, "failed")
		return nil, err
	}
	return &CreateShopOrderResult{
		ShopOrderID:           shopOrderID,
		Payment:               paymentResp,
		WalletAppliedCNYMinor: appliedMinor,
		PayableCNYMinor:       payableMinor,
	}, nil
}

// createPendingShopOrder 创建待支付订单，并按需用返点余额抵扣。
//
// 返回 (shopOrderID, payableMinor, appliedMinor)：
//   - appliedMinor 是实际抵扣金额（余额不足时按可用余额抵扣，可能为 0）；
//   - payableMinor 是还需外部支付的金额，为 0 表示全额抵扣。
func (s *ShopService) createPendingShopOrder(ctx context.Context, userID, productID int64, useWallet bool) (int64, int64, int64, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, 0, 0, err
	}
	defer func() { _ = tx.Rollback() }()
	var name, desc, imageURL, productType, grantRaw, status string
	var priceMinor, originalMinor int64
	var commissionBPS int
	var stock sql.NullInt64
	err = tx.QueryRowContext(ctx, `
SELECT name, description, image_url, product_type, price_cny_minor, original_price_cny_minor,
       grant_usd_amount::text, stock_quantity, commission_bps, status
FROM shop_products
WHERE tenant_id = 1 AND id = $1 AND deleted_at IS NULL
FOR UPDATE`, productID).Scan(&name, &desc, &imageURL, &productType, &priceMinor, &originalMinor, &grantRaw, &stock, &commissionBPS, &status)
	if errors.Is(err, sql.ErrNoRows) || status != "published" {
		return 0, 0, 0, infraerrors.NotFound("SHOP_PRODUCT_NOT_AVAILABLE", "product is not available")
	}
	if err != nil {
		return 0, 0, 0, err
	}
	if stock.Valid && stock.Int64 <= 0 {
		return 0, 0, 0, infraerrors.Conflict("SHOP_PRODUCT_SOLD_OUT", "product is sold out")
	}
	var shopOrderID int64
	err = tx.QueryRowContext(ctx, `
INSERT INTO shop_orders (tenant_id, user_id, product_id, snapshot_name, snapshot_description, snapshot_image_url,
    snapshot_product_type, snapshot_price_cny_minor, snapshot_grant_usd_amount, snapshot_commission_bps, payable_cny_minor)
VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $7)
RETURNING id`, userID, productID, name, desc, imageURL, productType, priceMinor, grantRaw, commissionBPS).Scan(&shopOrderID)
	if err != nil {
		return 0, 0, 0, err
	}
	// 先落单再抵扣：抵扣流水需要带上订单号作为幂等键。
	var appliedMinor int64
	if useWallet {
		programID, err := enabledShopWalletProgramIDTx(ctx, tx)
		if err != nil {
			return 0, 0, 0, err
		}
		appliedMinor, err = debitWalletForShopOrderTx(ctx, tx, programID, userID, shopOrderID, priceMinor)
		if err != nil {
			return 0, 0, 0, err
		}
		if appliedMinor > 0 {
			if _, err := tx.ExecContext(ctx, `UPDATE shop_orders SET wallet_applied_cny_minor = $2, payable_cny_minor = snapshot_price_cny_minor - $2, updated_at = NOW() WHERE id = $1`, shopOrderID, appliedMinor); err != nil {
				return 0, 0, 0, err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, 0, 0, err
	}
	return shopOrderID, priceMinor - appliedMinor, appliedMinor, nil
}

// shopWalletQueryer 让「查推广计划」既能跑在 *sql.Tx 上也能跑在 *sql.DB 上。
type shopWalletQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

// enabledShopWalletProgramIDTx 返回推广计划 program id；计划缺失或未启用时返回 0。
func enabledShopWalletProgramIDTx(ctx context.Context, q shopWalletQueryer) (int64, error) {
	var programID int64
	err := q.QueryRowContext(ctx, `SELECT id FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company' AND enabled = TRUE`).Scan(&programID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return programID, nil
}

// debitWalletForShopOrderTx 用人民币返点余额抵扣订单，返回实际抵扣金额。
//
// 余额不足时按可用余额抵扣（前端也会传 min(可用, 应付)，这里再兜一层，
// 因为锁行的顺序决定了必须以锁定后读到的余额为准）。
func debitWalletForShopOrderTx(ctx context.Context, tx *sql.Tx, programID, userID, orderID, wantMinor int64) (int64, error) {
	if programID <= 0 || wantMinor <= 0 {
		return 0, nil
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO distribution_cash_wallets (program_id, tenant_id, user_id) VALUES ($1, 1, $2) ON CONFLICT DO NOTHING`, programID, userID); err != nil {
		return 0, err
	}
	var available int64
	if err := tx.QueryRowContext(ctx, `SELECT available_cny_minor FROM distribution_cash_wallets WHERE program_id = $1 AND user_id = $2 FOR UPDATE`, programID, userID).Scan(&available); err != nil {
		return 0, err
	}
	applied := minInt64(available, wantMinor)
	if applied <= 0 {
		return 0, nil
	}
	if _, err := tx.ExecContext(ctx, `UPDATE distribution_cash_wallets SET available_cny_minor = available_cny_minor - $3, updated_at = NOW() WHERE program_id = $1 AND user_id = $2`, programID, userID, applied); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO distribution_wallet_ledger (program_id, tenant_id, user_id, action, amount_cny_minor, source_type, source_id, available_after, frozen_after, withdrawing_after, debt_after, idempotency_key, metadata)
SELECT $1, 1, $2, 'shop_wallet_deduction', $3, 'shop_order', $4, available_cny_minor, frozen_cny_minor, withdrawing_cny_minor, debt_cny_minor, $5, jsonb_build_object('label', '商城下单抵扣')
FROM distribution_cash_wallets WHERE program_id = $1 AND user_id = $2
ON CONFLICT DO NOTHING`, programID, userID, -applied, strconv.FormatInt(orderID, 10), fmt.Sprintf("shop:order:%d:wallet-deduction", orderID)); err != nil {
		return 0, err
	}
	return applied, nil
}

// refundShopOrderWalletTx 订单作废/退款时，把已抵扣的返点余额退回可用余额。
func refundShopOrderWalletTx(ctx context.Context, tx *sql.Tx, orderID int64) error {
	var userID, applied int64
	if err := tx.QueryRowContext(ctx, `SELECT user_id, wallet_applied_cny_minor FROM shop_orders WHERE tenant_id = 1 AND id = $1 FOR UPDATE`, orderID).Scan(&userID, &applied); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	if applied <= 0 {
		return nil
	}
	programID, err := enabledShopWalletProgramIDTx(ctx, tx)
	if err != nil {
		return err
	}
	if programID <= 0 {
		return nil
	}
	// 幂等：同一张订单只退一次。
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM distribution_wallet_ledger WHERE program_id = $1 AND idempotency_key = $2)`, programID, fmt.Sprintf("shop:order:%d:wallet-refund", orderID)).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO distribution_cash_wallets (program_id, tenant_id, user_id) VALUES ($1, 1, $2) ON CONFLICT DO NOTHING`, programID, userID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE distribution_cash_wallets SET available_cny_minor = available_cny_minor + $3, updated_at = NOW() WHERE program_id = $1 AND user_id = $2`, programID, userID, applied); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
INSERT INTO distribution_wallet_ledger (program_id, tenant_id, user_id, action, amount_cny_minor, source_type, source_id, available_after, frozen_after, withdrawing_after, debt_after, idempotency_key, metadata)
SELECT $1, 1, $2, 'shop_wallet_deduction_refund', $3, 'shop_order', $4, available_cny_minor, frozen_cny_minor, withdrawing_cny_minor, debt_cny_minor, $5, jsonb_build_object('label', '订单关闭退回抵扣')
FROM distribution_cash_wallets WHERE program_id = $1 AND user_id = $2
ON CONFLICT DO NOTHING`, programID, userID, applied, strconv.FormatInt(orderID, 10), fmt.Sprintf("shop:order:%d:wallet-refund", orderID))
	return err
}

// abortPendingShopOrder 把订单置为终止态并退回抵扣余额，供下单链路失败时调用。
func (s *ShopService) abortPendingShopOrder(ctx context.Context, shopOrderID int64, status string) {
	if status != "cancelled" && status != "failed" {
		status = "failed"
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		slog.Error("abort shop order: begin tx failed", "shop_order_id", shopOrderID, "error", err)
		return
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `UPDATE shop_orders SET status = $2, fulfillment_status = 'failed', updated_at = NOW() WHERE tenant_id = 1 AND id = $1 AND status = 'pending'`, shopOrderID, status); err != nil {
		slog.Error("abort shop order: update failed", "shop_order_id", shopOrderID, "error", err)
		return
	}
	if err := refundShopOrderWalletTx(ctx, tx, shopOrderID); err != nil {
		slog.Error("abort shop order: wallet refund failed", "shop_order_id", shopOrderID, "error", err)
		return
	}
	if err := tx.Commit(); err != nil {
		slog.Error("abort shop order: commit failed", "shop_order_id", shopOrderID, "error", err)
	}
}

// WalletBalance 返回当前用户可用于商城抵扣的返点余额。
func (s *ShopService) WalletBalance(ctx context.Context, userID int64) (*ShopWalletBalance, error) {
	programID, err := enabledShopWalletProgramIDTx(ctx, s.db)
	if err != nil {
		return nil, err
	}
	if programID <= 0 {
		return &ShopWalletBalance{}, nil
	}
	var available, frozen int64
	err = s.db.QueryRowContext(ctx, `SELECT available_cny_minor, frozen_cny_minor FROM distribution_cash_wallets WHERE program_id = $1 AND user_id = $2`, programID, userID).Scan(&available, &frozen)
	if errors.Is(err, sql.ErrNoRows) {
		return &ShopWalletBalance{Enabled: true}, nil
	}
	if err != nil {
		return nil, err
	}
	return &ShopWalletBalance{Enabled: true, AvailableCNYMinor: available, FrozenCNYMinor: frozen}, nil
}

func (s *ShopService) ValidatePendingOrderForPayment(ctx context.Context, shopOrderID, userID int64) (float64, error) {
	var priceMinor, payableMinor int64
	var status string
	err := s.db.QueryRowContext(ctx, `SELECT snapshot_price_cny_minor, payable_cny_minor, status FROM shop_orders WHERE tenant_id = 1 AND id = $1 AND user_id = $2`, shopOrderID, userID).Scan(&priceMinor, &payableMinor, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, infraerrors.NotFound("SHOP_ORDER_NOT_FOUND", "shop order not found")
	}
	if err != nil {
		return 0, err
	}
	if status != "pending" {
		return 0, infraerrors.Conflict("SHOP_ORDER_NOT_PENDING", "shop order is not pending")
	}
	// 抵扣过的订单只对外收「还需支付」的部分；待支付金额必须为正。
	if payableMinor <= 0 {
		if priceMinor <= 0 {
			return 0, infraerrors.BadRequest("SHOP_ORDER_PAYABLE_INVALID", "shop order payable amount is invalid")
		}
		payableMinor = priceMinor
	}
	return float64(payableMinor) / 100, nil
}

func (s *ShopService) AttachPaymentOrder(ctx context.Context, shopOrderID, userID, paymentOrderID int64) error {
	result, err := s.db.ExecContext(ctx, `UPDATE shop_orders SET payment_order_id = $3, updated_at = NOW() WHERE tenant_id = 1 AND id = $1 AND user_id = $2 AND status = 'pending' AND payment_order_id IS NULL`, shopOrderID, userID, paymentOrderID)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return infraerrors.Conflict("SHOP_ORDER_ATTACH_FAILED", "shop order cannot attach payment order")
	}
	return nil
}

// MarkPaymentOrderClosed 把未支付的商城订单置为终止态，并退回已抵扣的返点余额。
func (s *ShopService) MarkPaymentOrderClosed(ctx context.Context, paymentOrderID int64, status string) error {
	status = strings.TrimSpace(status)
	if status != "cancelled" && status != "failed" {
		return infraerrors.BadRequest("SHOP_ORDER_STATUS_INVALID", "shop order status is invalid")
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var shopOrderID int64
	err = tx.QueryRowContext(ctx, `SELECT id FROM shop_orders WHERE tenant_id = 1 AND payment_order_id = $1 AND status = 'pending' FOR UPDATE`, paymentOrderID).Scan(&shopOrderID)
	if errors.Is(err, sql.ErrNoRows) {
		// 订单已不是待支付状态：无需处理（可能已支付或已被关闭）。
		return nil
	}
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE shop_orders SET status = $2, fulfillment_status = 'failed', updated_at = NOW() WHERE id = $1`, shopOrderID, status); err != nil {
		return err
	}
	if err := refundShopOrderWalletTx(ctx, tx, shopOrderID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *ShopService) AdminFulfillOrder(ctx context.Context, orderID int64, note string) error {
	note = strings.TrimSpace(note)
	if note == "" {
		return infraerrors.BadRequest("INVALID_INPUT", "fulfillment note is required")
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var order ShopOrder
	var paymentOrderID sql.NullInt64
	var paidAt sql.NullTime
	var fulfilledAt sql.NullTime
	err = tx.QueryRowContext(ctx, `
SELECT id, user_id, product_id, payment_order_id, status, fulfillment_status, commission_status,
       snapshot_name, snapshot_description, snapshot_image_url, snapshot_product_type,
       snapshot_price_cny_minor, snapshot_grant_usd_amount::text, snapshot_commission_bps,
       fulfillment_note, paid_at, fulfilled_at
FROM shop_orders
WHERE tenant_id = 1 AND id = $1
FOR UPDATE`, orderID).Scan(&order.ID, &order.UserID, &order.ProductID, &paymentOrderID, &order.Status, &order.FulfillmentStatus, &order.CommissionStatus,
		&order.SnapshotName, &order.SnapshotDescription, &order.SnapshotImageURL, &order.SnapshotProductType,
		&order.SnapshotPriceCNYMinor, &order.SnapshotGrantUSD, &order.SnapshotCommissionBPS,
		&order.FulfillmentNote, &paidAt, &fulfilledAt)
	if errors.Is(err, sql.ErrNoRows) {
		return infraerrors.NotFound("SHOP_ORDER_NOT_FOUND", "shop order not found")
	}
	if err != nil {
		return err
	}
	if order.SnapshotProductType == ShopProductTypePlatformUSDBalance {
		return infraerrors.BadRequest("SHOP_ORDER_NOT_MANUAL", "platform balance products are fulfilled automatically")
	}
	if order.Status != "paid" || order.FulfillmentStatus != "pending" {
		return infraerrors.Conflict("SHOP_ORDER_NOT_READY", "shop order is not ready for manual fulfillment")
	}
	now := time.Now()
	if _, err := tx.ExecContext(ctx, `
UPDATE shop_orders
SET status = 'fulfilled',
    fulfillment_status = 'fulfilled',
    fulfillment_note = $2,
    fulfilled_at = COALESCE(fulfilled_at, $3),
    updated_at = NOW()
WHERE id = $1`, orderID, note, now); err != nil {
		return err
	}
	if paymentOrderID.Valid {
		if _, err := tx.ExecContext(ctx, `UPDATE payment_orders SET status = $2, completed_at = COALESCE(completed_at, NOW()), updated_at = NOW() WHERE id = $1`, paymentOrderID.Int64, OrderStatusCompleted); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *ShopService) FulfillPaidPaymentOrder(ctx context.Context, paymentOrderID int64) error {
	return s.fulfillShopOrder(ctx, paymentOrderID, 0)
}

// FulfillWalletPaidOrder 处理「全额返点余额抵扣」的订单：没有外部支付单，
// 但交付链路必须与正常支付完全一致（发货、佣金、库存扣减、幂等）。
func (s *ShopService) FulfillWalletPaidOrder(ctx context.Context, shopOrderID int64) error {
	return s.fulfillShopOrder(ctx, 0, shopOrderID)
}

// fulfillShopOrder 按 paymentOrderID 或 shopOrderID 二选一定位订单并完成交付。
// paymentOrderID <= 0 表示该订单没有外部支付单（全额余额抵扣）。
func (s *ShopService) fulfillShopOrder(ctx context.Context, paymentOrderID, shopOrderID int64) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var order ShopOrder
	var resolvedPaymentOrderID sql.NullInt64
	var paidAt sql.NullTime
	selector := `payment_order_id = $1`
	arg := paymentOrderID
	if paymentOrderID <= 0 {
		selector = `id = $1`
		arg = shopOrderID
	}
	err = tx.QueryRowContext(ctx, `
SELECT id, user_id, product_id, payment_order_id, status, fulfillment_status, commission_status, snapshot_name, snapshot_description,
       snapshot_image_url, snapshot_product_type, snapshot_price_cny_minor, wallet_applied_cny_minor, payable_cny_minor,
       snapshot_grant_usd_amount::text,
       snapshot_commission_bps, fulfillment_note, paid_at
FROM shop_orders
WHERE tenant_id = 1 AND `+selector+`
FOR UPDATE`, arg).Scan(&order.ID, &order.UserID, &order.ProductID, &resolvedPaymentOrderID, &order.Status, &order.FulfillmentStatus, &order.CommissionStatus,
		&order.SnapshotName, &order.SnapshotDescription, &order.SnapshotImageURL, &order.SnapshotProductType,
		&order.SnapshotPriceCNYMinor, &order.WalletAppliedCNYMinor, &order.PayableCNYMinor, &order.SnapshotGrantUSD,
		&order.SnapshotCommissionBPS, &order.FulfillmentNote, &paidAt)
	if errors.Is(err, sql.ErrNoRows) {
		return infraerrors.NotFound("SHOP_ORDER_NOT_FOUND", "shop order not found")
	}
	if err != nil {
		return err
	}
	if resolvedPaymentOrderID.Valid && resolvedPaymentOrderID.Int64 > 0 {
		paymentOrderID = resolvedPaymentOrderID.Int64
	} else {
		// 无外部支付单：全额抵扣订单，任何 payment_orders 回写都要跳过。
		paymentOrderID = 0
	}
	if order.SnapshotProductType == ShopProductTypeVirtual {
		if order.Status == "fulfilled" || (order.Status == "paid" && order.FulfillmentStatus == "pending") {
			if err := s.completeLinkedPaymentOrder(ctx, tx, paymentOrderID); err != nil {
				return err
			}
			return tx.Commit()
		}
		if order.Status != "pending" {
			return infraerrors.BadRequest("INVALID_STATUS", "order cannot fulfill in status "+order.Status)
		}
	}
	if order.FulfillmentStatus == "fulfilled" {
		if err := s.completeLinkedPaymentOrder(ctx, tx, paymentOrderID); err != nil {
			return err
		}
		return tx.Commit()
	}
	if order.SnapshotProductType == ShopProductTypePlatformUSDBalance {
		amount, parseErr := decimal.NewFromString(order.SnapshotGrantUSD)
		if parseErr != nil || !amount.IsPositive() {
			return infraerrors.BadRequest("SHOP_PRODUCT_GRANT_INVALID", "shop product grant amount is invalid")
		}
		if _, _, err := creditledger.Apply(ctx, tx, order.UserID, amount, creditctx.Metadata{
			EntryType: "shop_product_grant", SourceType: "shop_order", SourceID: strconv.FormatInt(order.ID, 10),
			IdempotencyKey: fmt.Sprintf("shop:order:%d:grant", order.ID), Transferable: false,
			Attributes: map[string]any{"payment_order_id": paymentOrderID, "product_id": order.ProductID},
		}, false); err != nil {
			return err
		}
	}
	if err := s.issueCommissionTx(ctx, tx, order); err != nil {
		return err
	}
	now := time.Now()
	result, err := tx.ExecContext(ctx, `UPDATE shop_products SET stock_quantity = CASE WHEN stock_quantity IS NULL THEN NULL ELSE stock_quantity - 1 END, sold_count = sold_count + 1, updated_at = NOW() WHERE id = $1 AND (stock_quantity IS NULL OR stock_quantity > 0)`, order.ProductID)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return infraerrors.Conflict("SHOP_PRODUCT_SOLD_OUT", "product is sold out")
	}
	if order.SnapshotProductType == ShopProductTypePlatformUSDBalance {
		if _, err := tx.ExecContext(ctx, `UPDATE shop_orders SET status = 'fulfilled', fulfillment_status = 'fulfilled', paid_at = COALESCE(paid_at, $2), fulfilled_at = COALESCE(fulfilled_at, $2), updated_at = NOW() WHERE id = $1`, order.ID, now); err != nil {
			return err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `UPDATE shop_orders SET status = 'paid', fulfillment_status = 'pending', paid_at = COALESCE(paid_at, $2), updated_at = NOW() WHERE id = $1`, order.ID, now); err != nil {
			return err
		}
	}
	if err := s.completeLinkedPaymentOrder(ctx, tx, paymentOrderID); err != nil {
		return err
	}
	return tx.Commit()
}

// completeLinkedPaymentOrder 把关联支付单回写为已完成。
// paymentOrderID <= 0 表示订单没有外部支付单（全额返点余额抵扣），直接跳过，
// 避免误写 payment_orders。本函数只负责写入，不再内部 Commit：
// 提交交给调用方统一处理，否则会出现「已提交事务再 Commit」的 sql.ErrTxDone。
func (s *ShopService) completeLinkedPaymentOrder(ctx context.Context, tx *sql.Tx, paymentOrderID int64) error {
	if paymentOrderID <= 0 {
		return nil
	}
	_, err := tx.ExecContext(ctx, `UPDATE payment_orders SET status = $2, completed_at = COALESCE(completed_at, NOW()), updated_at = NOW() WHERE id = $1 AND status <> $2`, paymentOrderID, OrderStatusCompleted)
	return err
}

func (s *ShopService) issueCommissionTx(ctx context.Context, tx *sql.Tx, order ShopOrder) error {
	// 返佣基数只取「外部实付金额」：用返点余额抵扣掉的那部分钱本身就是平台
	// 已经付出过的返点，再计一次会形成「返点 → 抵扣下单 → 再返点」的无锚增发。
	baseMinor := order.PayableCNYMinor
	if baseMinor <= 0 {
		baseMinor = order.SnapshotPriceCNYMinor - order.WalletAppliedCNYMinor
	}
	if order.SnapshotCommissionBPS <= 0 || baseMinor <= 0 {
		_, err := tx.ExecContext(ctx, `UPDATE shop_orders SET commission_status = 'none', updated_at = NOW() WHERE id = $1`, order.ID)
		return err
	}
	var inviterID sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT inviter_id FROM user_affiliates WHERE user_id = $1`, order.UserID).Scan(&inviterID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if !inviterID.Valid || inviterID.Int64 == order.UserID {
		_, err := tx.ExecContext(ctx, `UPDATE shop_orders SET commission_status = 'none', updated_at = NOW() WHERE id = $1`, order.ID)
		return err
	}
	var programID, freezeHours int64
	var enabled bool
	err := tx.QueryRowContext(ctx, `SELECT id, enabled, commission_freeze_hours FROM distribution_programs WHERE tenant_id = 1 AND code = 'compute_company' FOR SHARE`).Scan(&programID, &enabled, &freezeHours)
	if errors.Is(err, sql.ErrNoRows) || !enabled {
		_, err := tx.ExecContext(ctx, `UPDATE shop_orders SET commission_status = 'none', updated_at = NOW() WHERE id = $1`, order.ID)
		return err
	}
	if err != nil {
		return err
	}
	amountMinor := int64(math.Floor(float64(baseMinor*int64(order.SnapshotCommissionBPS))/10000 + 0.5))
	if amountMinor <= 0 {
		_, err := tx.ExecContext(ctx, `UPDATE shop_orders SET commission_status = 'none', updated_at = NOW() WHERE id = $1`, order.ID)
		return err
	}
	frozenUntil := time.Now().Add(time.Duration(freezeHours) * time.Hour)
	var commissionID int64
	err = tx.QueryRowContext(ctx, `
INSERT INTO shop_commission_records (tenant_id, shop_order_id, product_id, buyer_user_id, beneficiary_user_id,
    base_cny_minor, commission_bps, amount_cny_minor, frozen_until, idempotency_key)
VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (shop_order_id) DO NOTHING
RETURNING id`, order.ID, order.ProductID, order.UserID, inviterID.Int64, baseMinor, order.SnapshotCommissionBPS, amountMinor, frozenUntil, fmt.Sprintf("shop:commission:%d", order.ID)).Scan(&commissionID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO distribution_cash_wallets (program_id, tenant_id, user_id, frozen_cny_minor, lifetime_earned_cny_minor) VALUES ($1, 1, $2, $3, $3) ON CONFLICT (program_id, user_id) DO UPDATE SET frozen_cny_minor = distribution_cash_wallets.frozen_cny_minor + EXCLUDED.frozen_cny_minor, lifetime_earned_cny_minor = distribution_cash_wallets.lifetime_earned_cny_minor + EXCLUDED.lifetime_earned_cny_minor, updated_at = NOW()`, programID, inviterID.Int64, amountMinor); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO distribution_wallet_ledger (program_id, tenant_id, user_id, action, amount_cny_minor, source_type, source_id, available_after, frozen_after, withdrawing_after, debt_after, idempotency_key, metadata) SELECT $1, 1, $2, 'shop_commission_frozen', $3, 'shop_commission', $4, available_cny_minor, frozen_cny_minor, withdrawing_cny_minor, debt_cny_minor, $5, jsonb_build_object('label', '商城推广佣金') FROM distribution_cash_wallets WHERE program_id = $1 AND user_id = $2 ON CONFLICT DO NOTHING`, programID, inviterID.Int64, amountMinor, strconv.FormatInt(commissionID, 10), fmt.Sprintf("shop:commission:%d:wallet", order.ID)); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE shop_orders SET commission_status = 'frozen', updated_at = NOW() WHERE id = $1`, order.ID)
	return err
}

func (s *ShopService) ListMyOrders(ctx context.Context, userID int64, page, pageSize int) ([]ShopOrder, int64, error) {
	return s.listOrders(ctx, userID, page, pageSize)
}

func (s *ShopService) AdminListOrders(ctx context.Context, page, pageSize int) ([]ShopOrder, int64, error) {
	return s.listOrders(ctx, 0, page, pageSize)
}

func (s *ShopService) listOrders(ctx context.Context, userID int64, page, pageSize int) ([]ShopOrder, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	where := `o.tenant_id = 1`
	args := []any{}
	if userID > 0 {
		args = append(args, userID)
		where += fmt.Sprintf(" AND o.user_id = $%d", len(args))
	}
	var total int64
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM shop_orders o WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, pageSize, (page-1)*pageSize)
	rows, err := s.db.QueryContext(ctx, `
SELECT o.id, o.user_id, o.product_id, o.payment_order_id, o.status, o.fulfillment_status, o.commission_status,
       o.snapshot_name, o.snapshot_description, o.snapshot_image_url, o.snapshot_product_type,
       o.snapshot_price_cny_minor, o.wallet_applied_cny_minor, o.payable_cny_minor,
       o.snapshot_grant_usd_amount::text, o.snapshot_commission_bps,
       o.fulfillment_note, COALESCE(u.email, ''), COALESCE(o.order_no, ''), COALESCE(o.guest_token, ''),
       COALESCE(o.guest_contact, ''), COALESCE(o.snapshot_fulfillment_mode, 'manual'), COALESCE(o.delivery_hint, ''),
       o.delivery_submitted_at, COALESCE(o.rental_duration, ''), o.created_at, o.paid_at, o.fulfilled_at
FROM shop_orders o
LEFT JOIN users u ON u.id = o.user_id
WHERE `+where+`
ORDER BY o.created_at DESC
LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()
	out := []ShopOrder{}
	for rows.Next() {
		item, err := scanShopOrder(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, item)
	}
	return out, total, rows.Err()
}

func validateShopProductInput(in UpsertShopProductInput) error {
	if strings.TrimSpace(in.Name) == "" {
		return infraerrors.BadRequest("INVALID_INPUT", "product name is required")
	}
	if in.ProductType != ShopProductTypeVirtual && in.ProductType != ShopProductTypePlatformUSDBalance {
		return infraerrors.BadRequest("INVALID_INPUT", "invalid product type")
	}
	if in.PriceCNYMinor <= 0 {
		return infraerrors.BadRequest("INVALID_INPUT", "price must be greater than zero")
	}
	if in.ProductType == ShopProductTypePlatformUSDBalance {
		grant, err := decimal.NewFromString(in.GrantUSDAmount)
		if err != nil || !grant.IsPositive() {
			return infraerrors.BadRequest("INVALID_INPUT", "grant amount must be greater than zero")
		}
	}
	if in.CommissionBPS < 0 || in.CommissionBPS > 10000 {
		return infraerrors.BadRequest("INVALID_INPUT", "commission percent is invalid")
	}
	if in.Status != "draft" && in.Status != "published" && in.Status != "archived" {
		return infraerrors.BadRequest("INVALID_INPUT", "invalid product status")
	}
	return nil
}

func normalizeShopProductInput(in UpsertShopProductInput) UpsertShopProductInput {
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	in.ImageURL = strings.TrimSpace(in.ImageURL)
	in.ProductType = strings.TrimSpace(in.ProductType)
	if in.ProductType == "" {
		in.ProductType = ShopProductTypeVirtual
	}
	in.Status = strings.TrimSpace(in.Status)
	if in.Status == "" {
		in.Status = "draft"
	}
	in.FulfillmentMode = strings.TrimSpace(in.FulfillmentMode)
	if !isKnownFulfillmentMode(in.FulfillmentMode) {
		in.FulfillmentMode = ShopFulfillmentManual
	}
	in.DeliveryFormHint = strings.TrimSpace(in.DeliveryFormHint)
	in.BadgeText = strings.TrimSpace(in.BadgeText)
	in.SpecLabel = strings.TrimSpace(in.SpecLabel)
	in.Category = strings.TrimSpace(in.Category)
	// 品类集合由 shop_categories 表定义（迁移 208），这里只做「空值兜底」，
	// 具体是否存在交由 Create/Update 查库校验，否则后台新建的品类会被静默改成 other。
	if in.Category == "" {
		in.Category = ShopCategoryOther
	}
	in.Gallery = normalizeGallery(in.Gallery)
	in.GrantUSDAmount = normalizeDecimalString(in.GrantUSDAmount)
	return in
}

// normalizeGallery 去掉空串与重复项，并限制图廊长度，避免后台误填导致前端渲染失控。
func normalizeGallery(items []string) []string {
	const maxGalleryItems = 6
	clean := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		clean = append(clean, trimmed)
	}
	if len(clean) > maxGalleryItems {
		clean = clean[:maxGalleryItems]
	}
	return clean
}

func normalizeDecimalString(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "0"
	}
	if _, err := decimal.NewFromString(value); err != nil {
		return "0"
	}
	return value
}

type productScanner interface {
	Scan(dest ...any) error
}

func scanShopProduct(rows productScanner) (ShopProduct, error) {
	var item ShopProduct
	var stock sql.NullInt64
	var created, updated time.Time
	var galleryRaw string
	err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.ImageURL, &item.ProductType, &item.PriceCNYMinor,
		&item.OriginalPriceCNYMinor, &item.GrantUSDAmount, &stock, &item.SoldCount, &item.CommissionBPS,
		&item.Status, &item.SortOrder, &item.FulfillmentMode, &item.DeliveryFormHint, &item.BadgeText,
		&item.SpecLabel, &item.Highlight, &item.Category, &galleryRaw, &created, &updated)
	if stock.Valid {
		item.StockQuantity = &stock.Int64
	}
	item.Gallery = decodeGalleryJSON(galleryRaw)
	if item.Category == "" {
		item.Category = ShopCategoryOther
	}
	item.CreatedAt = created.Format(time.RFC3339)
	item.UpdatedAt = updated.Format(time.RFC3339)
	return item, err
}

// decodeGalleryJSON 把 gallery_json 列解析为图片 URL 数组。
// 空值或脏数据一律退化为空数组，避免把无效 JSON 透传给前端。
func decodeGalleryJSON(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal([]byte(trimmed), &out); err != nil || out == nil {
		return []string{}
	}
	return out
}

// encodeGalleryJSON 把图片 URL 数组序列化为 gallery_json 列值。
func encodeGalleryJSON(items []string) string {
	clean := make([]string, 0, len(items))
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			clean = append(clean, trimmed)
		}
	}
	encoded, err := json.Marshal(clean)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}

func scanShopOrder(rows productScanner) (ShopOrder, error) {
	var item ShopOrder
	var paymentOrderID sql.NullInt64
	var created time.Time
	var paidAt, fulfilledAt, deliverySubmittedAt sql.NullTime
	err := rows.Scan(&item.ID, &item.UserID, &item.ProductID, &paymentOrderID, &item.Status, &item.FulfillmentStatus,
		&item.CommissionStatus, &item.SnapshotName, &item.SnapshotDescription, &item.SnapshotImageURL,
		&item.SnapshotProductType, &item.SnapshotPriceCNYMinor, &item.WalletAppliedCNYMinor, &item.PayableCNYMinor,
		&item.SnapshotGrantUSD,
		&item.SnapshotCommissionBPS, &item.FulfillmentNote, &item.UserEmail, &item.OrderNo, &item.GuestToken,
		&item.GuestContact, &item.SnapshotFulfillmentMode, &item.DeliveryHint, &deliverySubmittedAt,
		&item.RentalDuration, &created, &paidAt, &fulfilledAt)
	if err != nil {
		return item, err
	}
	if paymentOrderID.Valid {
		item.PaymentOrderID = &paymentOrderID.Int64
	}
	item.CreatedAt = created.Format(time.RFC3339)
	if deliverySubmittedAt.Valid {
		v := deliverySubmittedAt.Time.Format(time.RFC3339)
		item.DeliverySubmittedAt = &v
	}
	if paidAt.Valid {
		v := paidAt.Time.Format(time.RFC3339)
		item.PaidAt = &v
	}
	if fulfilledAt.Valid {
		v := fulfilledAt.Time.Format(time.RFC3339)
		item.FulfilledAt = &v
	}
	return item, nil
}
