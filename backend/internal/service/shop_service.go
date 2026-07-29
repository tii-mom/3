package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
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

type ShopService struct {
	db             *sql.DB
	paymentService *PaymentService
}

func NewShopService(db *sql.DB) *ShopService {
	return &ShopService{db: db}
}

func (s *ShopService) SetPaymentService(paymentService *PaymentService) {
	s.paymentService = paymentService
}

type ShopProduct struct {
	ID                    int64  `json:"id"`
	Name                  string `json:"name"`
	Description           string `json:"description"`
	ImageURL              string `json:"image_url"`
	ProductType           string `json:"product_type"`
	PriceCNYMinor         int64  `json:"price_cny_minor"`
	OriginalPriceCNYMinor int64  `json:"original_price_cny_minor"`
	GrantUSDAmount        string `json:"grant_usd_amount"`
	StockQuantity         *int64 `json:"stock_quantity,omitempty"`
	SoldCount             int64  `json:"sold_count"`
	CommissionBPS         int    `json:"commission_bps"`
	Status                string `json:"status"`
	SortOrder             int    `json:"sort_order"`
	CreatedAt             string `json:"created_at,omitempty"`
	UpdatedAt             string `json:"updated_at,omitempty"`
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
	ID                    int64   `json:"id"`
	UserID                int64   `json:"user_id"`
	ProductID             int64   `json:"product_id"`
	PaymentOrderID        *int64  `json:"payment_order_id,omitempty"`
	Status                string  `json:"status"`
	FulfillmentStatus     string  `json:"fulfillment_status"`
	CommissionStatus      string  `json:"commission_status"`
	SnapshotName          string  `json:"snapshot_name"`
	SnapshotDescription   string  `json:"snapshot_description"`
	SnapshotImageURL      string  `json:"snapshot_image_url"`
	SnapshotProductType   string  `json:"snapshot_product_type"`
	SnapshotPriceCNYMinor int64   `json:"snapshot_price_cny_minor"`
	SnapshotGrantUSD      string  `json:"snapshot_grant_usd_amount"`
	SnapshotCommissionBPS int     `json:"snapshot_commission_bps"`
	UserEmail             string  `json:"user_email,omitempty"`
	CreatedAt             string  `json:"created_at"`
	PaidAt                *string `json:"paid_at,omitempty"`
	FulfilledAt           *string `json:"fulfilled_at,omitempty"`
}

type CreateShopOrderResult struct {
	ShopOrderID int64                `json:"shop_order_id"`
	Payment     *CreateOrderResponse `json:"payment"`
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
	where := `tenant_id = 1 AND deleted_at IS NULL`
	if !admin {
		where += ` AND status = 'published'`
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, name, description, image_url, product_type, price_cny_minor, original_price_cny_minor,
       grant_usd_amount::text, stock_quantity, sold_count, commission_bps, status, sort_order,
       created_at, updated_at
FROM shop_products
WHERE `+where+`
ORDER BY sort_order ASC, id DESC`)
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
	var item ShopProduct
	row := s.db.QueryRowContext(ctx, `
INSERT INTO shop_products (tenant_id, name, description, image_url, product_type, price_cny_minor,
    original_price_cny_minor, grant_usd_amount, stock_quantity, commission_bps, status, sort_order)
VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, name, description, image_url, product_type, price_cny_minor, original_price_cny_minor,
       grant_usd_amount::text, stock_quantity, sold_count, commission_bps, status, sort_order, created_at, updated_at`,
		in.Name, in.Description, in.ImageURL, in.ProductType, in.PriceCNYMinor, in.OriginalPriceCNYMinor,
		normalizeDecimalString(in.GrantUSDAmount), in.StockQuantity, in.CommissionBPS, in.Status, in.SortOrder)
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
	var item ShopProduct
	row := s.db.QueryRowContext(ctx, `
UPDATE shop_products
SET name = $2, description = $3, image_url = $4, product_type = $5, price_cny_minor = $6,
    original_price_cny_minor = $7, grant_usd_amount = $8, stock_quantity = $9, commission_bps = $10,
    status = $11, sort_order = $12, updated_at = NOW()
WHERE tenant_id = 1 AND id = $1 AND deleted_at IS NULL
RETURNING id, name, description, image_url, product_type, price_cny_minor, original_price_cny_minor,
       grant_usd_amount::text, stock_quantity, sold_count, commission_bps, status, sort_order, created_at, updated_at`,
		id, in.Name, in.Description, in.ImageURL, in.ProductType, in.PriceCNYMinor, in.OriginalPriceCNYMinor,
		normalizeDecimalString(in.GrantUSDAmount), in.StockQuantity, in.CommissionBPS, in.Status, in.SortOrder)
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

func (s *ShopService) CreateOrderAndPayment(ctx context.Context, userID int64, productID int64, paymentType, returnURL, clientIP, srcHost, srcURL, locale string, isMobile, isWeChatBrowser bool, openID string) (*CreateShopOrderResult, error) {
	if s.paymentService == nil {
		return nil, infraerrors.Forbidden("PAYMENT_UNAVAILABLE", "payment system is unavailable")
	}
	shopOrderID, amount, err := s.createPendingShopOrder(ctx, userID, productID)
	if err != nil {
		return nil, err
	}
	paymentResp, err := s.paymentService.CreateOrder(ctx, CreateOrderRequest{
		UserID:          userID,
		Amount:          amount,
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
		_, _ = s.db.ExecContext(ctx, `UPDATE shop_orders SET status = 'failed', updated_at = NOW() WHERE id = $1 AND status = 'pending'`, shopOrderID)
		return nil, err
	}
	return &CreateShopOrderResult{ShopOrderID: shopOrderID, Payment: paymentResp}, nil
}

func (s *ShopService) createPendingShopOrder(ctx context.Context, userID, productID int64) (int64, float64, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, 0, err
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
		return 0, 0, infraerrors.NotFound("SHOP_PRODUCT_NOT_AVAILABLE", "product is not available")
	}
	if err != nil {
		return 0, 0, err
	}
	if stock.Valid && stock.Int64 <= 0 {
		return 0, 0, infraerrors.Conflict("SHOP_PRODUCT_SOLD_OUT", "product is sold out")
	}
	var shopOrderID int64
	err = tx.QueryRowContext(ctx, `
INSERT INTO shop_orders (tenant_id, user_id, product_id, snapshot_name, snapshot_description, snapshot_image_url,
    snapshot_product_type, snapshot_price_cny_minor, snapshot_grant_usd_amount, snapshot_commission_bps)
VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id`, userID, productID, name, desc, imageURL, productType, priceMinor, grantRaw, commissionBPS).Scan(&shopOrderID)
	if err != nil {
		return 0, 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}
	return shopOrderID, float64(priceMinor) / 100, nil
}

func (s *ShopService) ValidatePendingOrderForPayment(ctx context.Context, shopOrderID, userID int64) (float64, error) {
	var priceMinor int64
	var status string
	err := s.db.QueryRowContext(ctx, `SELECT snapshot_price_cny_minor, status FROM shop_orders WHERE tenant_id = 1 AND id = $1 AND user_id = $2`, shopOrderID, userID).Scan(&priceMinor, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, infraerrors.NotFound("SHOP_ORDER_NOT_FOUND", "shop order not found")
	}
	if err != nil {
		return 0, err
	}
	if status != "pending" {
		return 0, infraerrors.Conflict("SHOP_ORDER_NOT_PENDING", "shop order is not pending")
	}
	return float64(priceMinor) / 100, nil
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

func (s *ShopService) MarkPaymentOrderClosed(ctx context.Context, paymentOrderID int64, status string) error {
	status = strings.TrimSpace(status)
	if status != "cancelled" && status != "failed" {
		return infraerrors.BadRequest("SHOP_ORDER_STATUS_INVALID", "shop order status is invalid")
	}
	_, err := s.db.ExecContext(ctx, `
UPDATE shop_orders
SET status = $2, fulfillment_status = 'failed', updated_at = NOW()
WHERE tenant_id = 1
  AND payment_order_id = $1
  AND status = 'pending'`, paymentOrderID, status)
	return err
}

func (s *ShopService) FulfillPaidPaymentOrder(ctx context.Context, paymentOrderID int64) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var order ShopOrder
	var paidAt sql.NullTime
	err = tx.QueryRowContext(ctx, `
SELECT id, user_id, product_id, status, fulfillment_status, commission_status, snapshot_name, snapshot_description,
       snapshot_image_url, snapshot_product_type, snapshot_price_cny_minor, snapshot_grant_usd_amount::text,
       snapshot_commission_bps, paid_at
FROM shop_orders
WHERE tenant_id = 1 AND payment_order_id = $1
FOR UPDATE`, paymentOrderID).Scan(&order.ID, &order.UserID, &order.ProductID, &order.Status, &order.FulfillmentStatus, &order.CommissionStatus,
		&order.SnapshotName, &order.SnapshotDescription, &order.SnapshotImageURL, &order.SnapshotProductType,
		&order.SnapshotPriceCNYMinor, &order.SnapshotGrantUSD, &order.SnapshotCommissionBPS, &paidAt)
	if errors.Is(err, sql.ErrNoRows) {
		return infraerrors.NotFound("SHOP_ORDER_NOT_FOUND", "shop order not found")
	}
	if err != nil {
		return err
	}
	if order.FulfillmentStatus == "fulfilled" {
		_, err = tx.ExecContext(ctx, `UPDATE payment_orders SET status = $2, completed_at = COALESCE(completed_at, NOW()), updated_at = NOW() WHERE id = $1 AND status <> $2`, paymentOrderID, OrderStatusCompleted)
		if err != nil {
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
	if _, err := tx.ExecContext(ctx, `UPDATE shop_orders SET status = 'fulfilled', fulfillment_status = 'fulfilled', paid_at = COALESCE(paid_at, $2), fulfilled_at = COALESCE(fulfilled_at, $2), updated_at = NOW() WHERE id = $1`, order.ID, now); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE payment_orders SET status = $2, completed_at = COALESCE(completed_at, NOW()), updated_at = NOW() WHERE id = $1`, paymentOrderID, OrderStatusCompleted); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *ShopService) issueCommissionTx(ctx context.Context, tx *sql.Tx, order ShopOrder) error {
	if order.SnapshotCommissionBPS <= 0 || order.SnapshotPriceCNYMinor <= 0 {
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
	amountMinor := int64(math.Floor(float64(order.SnapshotPriceCNYMinor*int64(order.SnapshotCommissionBPS))/10000 + 0.5))
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
RETURNING id`, order.ID, order.ProductID, order.UserID, inviterID.Int64, order.SnapshotPriceCNYMinor, order.SnapshotCommissionBPS, amountMinor, frozenUntil, fmt.Sprintf("shop:commission:%d", order.ID)).Scan(&commissionID)
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
       o.snapshot_price_cny_minor, o.snapshot_grant_usd_amount::text, o.snapshot_commission_bps,
       COALESCE(u.email, ''), o.created_at, o.paid_at, o.fulfilled_at
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
	in.GrantUSDAmount = normalizeDecimalString(in.GrantUSDAmount)
	return in
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
	err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.ImageURL, &item.ProductType, &item.PriceCNYMinor,
		&item.OriginalPriceCNYMinor, &item.GrantUSDAmount, &stock, &item.SoldCount, &item.CommissionBPS,
		&item.Status, &item.SortOrder, &created, &updated)
	if stock.Valid {
		item.StockQuantity = &stock.Int64
	}
	item.CreatedAt = created.Format(time.RFC3339)
	item.UpdatedAt = updated.Format(time.RFC3339)
	return item, err
}

func scanShopOrder(rows productScanner) (ShopOrder, error) {
	var item ShopOrder
	var paymentOrderID sql.NullInt64
	var created time.Time
	var paidAt, fulfilledAt sql.NullTime
	err := rows.Scan(&item.ID, &item.UserID, &item.ProductID, &paymentOrderID, &item.Status, &item.FulfillmentStatus,
		&item.CommissionStatus, &item.SnapshotName, &item.SnapshotDescription, &item.SnapshotImageURL,
		&item.SnapshotProductType, &item.SnapshotPriceCNYMinor, &item.SnapshotGrantUSD,
		&item.SnapshotCommissionBPS, &item.UserEmail, &created, &paidAt, &fulfilledAt)
	if err != nil {
		return item, err
	}
	if paymentOrderID.Valid {
		item.PaymentOrderID = &paymentOrderID.Int64
	}
	item.CreatedAt = created.Format(time.RFC3339)
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
