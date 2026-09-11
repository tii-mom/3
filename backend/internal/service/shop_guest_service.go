package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// 本文件承载「免登录直充」链路：游客下单、支付后补交交付资料、凭订单号查单。
// 订单统一挂在系统游客账号下，通过 guest_token + guest_contact 做归属与查单凭据。

const (
	guestOrderNoPrefix    = "C"
	guestOrderNoRandomLen = 6
	guestTokenBytes       = 32

	// 限流窗口与阈值（进程内计数，防刷单，重启后重置）
	guestLimitWindow     = 10 * time.Minute
	guestLimitPerIP      = 20
	guestLimitPerContact = 5
)

// CreateGuestOrderResult 免登录下单结果，guest_token 仅此一次返回，前端需本地保存用于查单。
type CreateGuestOrderResult struct {
	ShopOrderID int64                `json:"shop_order_id"`
	OrderNo     string               `json:"order_no"`
	GuestToken  string               `json:"guest_token"`
	Payment     *CreateOrderResponse `json:"payment"`
}

// GuestOrderLookup 查单返回体，永不返回 delivery_payload 原文。
type GuestOrderLookup struct {
	OrderNo               string  `json:"order_no"`
	Status                string  `json:"status"`
	FulfillmentStatus     string  `json:"fulfillment_status"`
	SnapshotName          string  `json:"snapshot_name"`
	SnapshotDescription   string  `json:"snapshot_description"`
	SnapshotImageURL      string  `json:"snapshot_image_url"`
	SnapshotPriceCNYMinor int64   `json:"snapshot_price_cny_minor"`
	SnapshotFulfillment   string  `json:"snapshot_fulfillment_mode"`
	FulfillmentNote       string  `json:"fulfillment_note"`
	DeliveryHint          string  `json:"delivery_hint,omitempty"`
	DeliverySubmitted     bool    `json:"delivery_submitted"`
	RentalDuration        string  `json:"rental_duration,omitempty"`
	CreatedAt             string  `json:"created_at"`
	PaidAt                *string `json:"paid_at,omitempty"`
}

type guestOrderLimiter struct {
	mu      sync.Mutex
	records map[string][]time.Time
}

func newGuestOrderLimiter() *guestOrderLimiter {
	return &guestOrderLimiter{records: map[string][]time.Time{}}
}

func (l *guestOrderLimiter) allow(key string, limit int) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	kept := l.records[key][:0]
	for _, ts := range l.records[key] {
		if now.Sub(ts) < guestLimitWindow {
			kept = append(kept, ts)
		}
	}
	allowed := len(kept) < limit
	if allowed {
		kept = append(kept, now)
	}
	if len(kept) == 0 {
		delete(l.records, key)
	} else {
		l.records[key] = kept
	}
	return allowed
}

// CreateGuestOrderAndPayment 免登录下单：校验联系信息 → 创建游客订单 → 拉起支付。
func (s *ShopService) CreateGuestOrderAndPayment(
	ctx context.Context,
	contact string,
	productID int64,
	paymentType string,
	returnURL string,
	clientIP string,
	srcHost string,
	srcURL string,
	locale string,
	isMobile bool,
	isWeChatBrowser bool,
	openID string,
) (*CreateGuestOrderResult, error) {
	if s.paymentService == nil {
		return nil, infraerrors.Forbidden("PAYMENT_UNAVAILABLE", "payment system is unavailable")
	}
	contact = normalizeGuestContact(contact)
	if err := validateGuestContact(contact); err != nil {
		return nil, err
	}
	if s.guestLimiter != nil {
		if clientIP != "" && !s.guestLimiter.allow("ip:"+clientIP, guestLimitPerIP) {
			return nil, infraerrors.TooManyRequests("GUEST_ORDER_RATE_LIMITED", "too many orders from this network, please try again later")
		}
		if !s.guestLimiter.allow("contact:"+contact, guestLimitPerContact) {
			return nil, infraerrors.TooManyRequests("GUEST_ORDER_RATE_LIMITED", "too many orders for this contact, please try again later")
		}
	}

	guestUserID, err := s.resolveGuestUserID(ctx)
	if err != nil {
		return nil, err
	}
	shopOrderID, orderNo, guestToken, amount, err := s.createPendingGuestOrder(ctx, guestUserID, contact, productID)
	if err != nil {
		return nil, err
	}
	paymentResp, err := s.paymentService.CreateOrder(ctx, CreateOrderRequest{
		UserID:          guestUserID,
		Amount:          amount,
		PaymentType:     paymentType,
		OpenID:          openID,
		ClientIP:        clientIP,
		IsMobile:        isMobile,
		IsWeChatBrowser: isWeChatBrowser,
		SrcHost:         srcHost,
		SrcURL:          srcURL,
		ReturnURL:       returnURL,
		PaymentSource:   "shop_guest",
		OrderType:       "shop",
		ShopOrderID:     shopOrderID,
		Locale:          locale,
	})
	if err != nil {
		_, _ = s.db.ExecContext(ctx, `UPDATE shop_orders SET status = 'failed', updated_at = NOW() WHERE id = $1 AND status = 'pending'`, shopOrderID)
		return nil, err
	}
	return &CreateGuestOrderResult{
		ShopOrderID: shopOrderID,
		OrderNo:     orderNo,
		GuestToken:  guestToken,
		Payment:     paymentResp,
	}, nil
}

func (s *ShopService) resolveGuestUserID(ctx context.Context) (int64, error) {
	var userID int64
	err := s.db.QueryRowContext(ctx, `SELECT id FROM users WHERE email = $1 AND deleted_at IS NULL`, shopGuestUserEmail).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, infraerrors.InternalServer("GUEST_ACCOUNT_MISSING", "guest account is not initialized")
	}
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (s *ShopService) createPendingGuestOrder(ctx context.Context, guestUserID int64, contact string, productID int64) (int64, string, string, float64, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, "", "", 0, err
	}
	defer func() { _ = tx.Rollback() }()

	var name, desc, imageURL, productType, grantRaw, status, fulfillmentMode string
	var priceMinor, originalMinor int64
	var commissionBPS int
	var stock sql.NullInt64
	err = tx.QueryRowContext(ctx, `
SELECT name, description, image_url, product_type, price_cny_minor, original_price_cny_minor,
       grant_usd_amount::text, stock_quantity, commission_bps, status, fulfillment_mode
FROM shop_products
WHERE tenant_id = 1 AND id = $1 AND deleted_at IS NULL
FOR UPDATE`, productID).Scan(&name, &desc, &imageURL, &productType, &priceMinor, &originalMinor, &grantRaw, &stock, &commissionBPS, &status, &fulfillmentMode)
	if errors.Is(err, sql.ErrNoRows) || status != "published" {
		return 0, "", "", 0, infraerrors.NotFound("SHOP_PRODUCT_NOT_AVAILABLE", "product is not available")
	}
	if err != nil {
		return 0, "", "", 0, err
	}
	if stock.Valid && stock.Int64 <= 0 {
		return 0, "", "", 0, infraerrors.Conflict("SHOP_PRODUCT_SOLD_OUT", "product is sold out")
	}

	orderNo, err := generateGuestOrderNo(ctx, tx)
	if err != nil {
		return 0, "", "", 0, err
	}
	guestToken, err := randomToken(guestTokenBytes)
	if err != nil {
		return 0, "", "", 0, err
	}

	var shopOrderID int64
	err = tx.QueryRowContext(ctx, `
INSERT INTO shop_orders (tenant_id, user_id, product_id, snapshot_name, snapshot_description, snapshot_image_url,
    snapshot_product_type, snapshot_price_cny_minor, snapshot_grant_usd_amount, snapshot_commission_bps,
    order_no, guest_token, guest_contact, snapshot_fulfillment_mode)
VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING id`, guestUserID, productID, name, desc, imageURL, productType, priceMinor, grantRaw, commissionBPS,
		orderNo, guestToken, contact, fulfillmentMode).Scan(&shopOrderID)
	if err != nil {
		return 0, "", "", 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, "", "", 0, err
	}
	return shopOrderID, orderNo, guestToken, float64(priceMinor) / 100, nil
}

// SubmitGuestDeliveryInfo 支付后（或待支付时）补交交付资料。
// payload 会被加密存储；session_topup 模式会尝试从整段 JSON 中提取 accessToken。
func (s *ShopService) SubmitGuestDeliveryInfo(ctx context.Context, orderNo string, contact string, payload string, rentalDuration string) error {
	if s.encryptor == nil {
		return infraerrors.InternalServer("ENCRYPTOR_UNAVAILABLE", "delivery encryption is not configured")
	}
	orderNo = strings.TrimSpace(orderNo)
	contact = normalizeGuestContact(contact)
	if orderNo == "" {
		return infraerrors.BadRequest("INVALID_INPUT", "order number is required")
	}
	if err := validateGuestContact(contact); err != nil {
		return err
	}
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return infraerrors.BadRequest("INVALID_INPUT", "delivery info is required")
	}
	if len(payload) > 20000 {
		return infraerrors.BadRequest("INVALID_INPUT", "delivery info is too long")
	}

	var id int64
	var status, mode string
	err := s.db.QueryRowContext(ctx, `
SELECT id, status, COALESCE(snapshot_fulfillment_mode, 'manual')
FROM shop_orders
WHERE tenant_id = 1 AND order_no = $1 AND guest_contact = $2`, orderNo, contact).Scan(&id, &status, &mode)
	if errors.Is(err, sql.ErrNoRows) {
		return infraerrors.NotFound("SHOP_ORDER_NOT_FOUND", "order not found, please check order number and contact")
	}
	if err != nil {
		return err
	}
	if status == "cancelled" || status == "refunded" || status == "failed" {
		return infraerrors.Conflict("SHOP_ORDER_CLOSED", "order is closed and cannot be updated")
	}

	normalized := payload
	if mode == ShopFulfillmentSessionTopup {
		token, err := extractSessionToken(payload)
		if err != nil {
			return err
		}
		normalized = token
	}

	encrypted, err := s.encryptor.Encrypt(normalized)
	if err != nil {
		return infraerrors.InternalServer("ENCRYPT_FAILED", "failed to secure delivery info")
	}
	_, err = s.db.ExecContext(ctx, `
UPDATE shop_orders
SET delivery_payload = $2, delivery_hint = $3, delivery_submitted_at = NOW(),
    rental_duration = COALESCE(NULLIF($4, ''), rental_duration), updated_at = NOW()
WHERE id = $1`, id, encrypted, buildDeliveryHint(mode, normalized, contact), strings.TrimSpace(rentalDuration))
	return err
}

// LookupGuestOrder 凭订单号 + 联系方式查单，不返回任何敏感字段。
func (s *ShopService) LookupGuestOrder(ctx context.Context, orderNo string, contact string) (*GuestOrderLookup, error) {
	orderNo = strings.TrimSpace(orderNo)
	contact = normalizeGuestContact(contact)
	if orderNo == "" {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "order number is required")
	}
	if err := validateGuestContact(contact); err != nil {
		return nil, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT id, status, fulfillment_status, snapshot_name, snapshot_description, snapshot_image_url,
       snapshot_price_cny_minor, COALESCE(snapshot_fulfillment_mode, 'manual'), fulfillment_note,
       COALESCE(delivery_hint, ''), delivery_submitted_at, COALESCE(rental_duration, ''), created_at, paid_at
FROM shop_orders
WHERE tenant_id = 1 AND order_no = $1 AND guest_contact = $2`, orderNo, contact)
	var (
		id                   int64
		status, fStatus      string
		name, desc, imageURL string
		priceMinor           int64
		mode, note, hint     string
		deliveryAt           sql.NullTime
		rental               string
		created              time.Time
		paidAt               sql.NullTime
	)
	if err := row.Scan(&id, &status, &fStatus, &name, &desc, &imageURL, &priceMinor, &mode, &note, &hint, &deliveryAt, &rental, &created, &paidAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, infraerrors.NotFound("SHOP_ORDER_NOT_FOUND", "order not found, please check order number and contact")
		}
		return nil, err
	}
	out := &GuestOrderLookup{
		OrderNo:               orderNo,
		Status:                status,
		FulfillmentStatus:     fStatus,
		SnapshotName:          name,
		SnapshotDescription:   desc,
		SnapshotImageURL:      imageURL,
		SnapshotPriceCNYMinor: priceMinor,
		SnapshotFulfillment:   mode,
		FulfillmentNote:       note,
		DeliveryHint:          hint,
		DeliverySubmitted:     deliveryAt.Valid,
		RentalDuration:        rental,
		CreatedAt:             created.Format(time.RFC3339),
	}
	if paidAt.Valid {
		v := paidAt.Time.Format(time.RFC3339)
		out.PaidAt = &v
	}
	return out, nil
}

// AdminDecryptDelivery 后台发货时解密交付资料，仅供管理员调用。
func (s *ShopService) AdminDecryptDelivery(ctx context.Context, orderID int64) (string, error) {
	if s.encryptor == nil {
		return "", infraerrors.InternalServer("ENCRYPTOR_UNAVAILABLE", "delivery encryption is not configured")
	}
	var payload string
	err := s.db.QueryRowContext(ctx, `SELECT delivery_payload FROM shop_orders WHERE tenant_id = 1 AND id = $1`, orderID).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return "", infraerrors.NotFound("SHOP_ORDER_NOT_FOUND", "order not found")
	}
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(payload) == "" {
		return "", nil
	}
	plain, err := s.encryptor.Decrypt(payload)
	if err != nil {
		return "", infraerrors.InternalServer("DECRYPT_FAILED", "failed to decrypt delivery info")
	}
	return plain, nil
}

// --- helpers ---

func normalizeGuestContact(contact string) string {
	return strings.ToLower(strings.TrimSpace(contact))
}

func validateGuestContact(contact string) error {
	if contact == "" {
		return infraerrors.BadRequest("INVALID_INPUT", "contact is required")
	}
	if len(contact) > 120 {
		return infraerrors.BadRequest("INVALID_INPUT", "contact is too long")
	}
	// 允许邮箱或手机号（含国际区号），不做过严校验以免误杀
	if strings.Contains(contact, "@") {
		if !strings.Contains(contact, ".") || strings.HasPrefix(contact, "@") || strings.HasSuffix(contact, "@") {
			return infraerrors.BadRequest("INVALID_INPUT", "email format is invalid")
		}
		return nil
	}
	for _, r := range contact {
		if (r >= '0' && r <= '9') || r == '+' || r == '-' || r == ' ' {
			continue
		}
		return infraerrors.BadRequest("INVALID_INPUT", "phone number format is invalid")
	}
	if len(contact) < 6 {
		return infraerrors.BadRequest("INVALID_INPUT", "phone number format is invalid")
	}
	return nil
}

// extractSessionToken 从用户粘贴的内容里取出 accessToken：
// 支持整段 session JSON，也支持直接粘贴的 token。
func extractSessionToken(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", infraerrors.BadRequest("INVALID_INPUT", "session is empty")
	}
	if strings.HasPrefix(raw, "{") {
		var payload map[string]any
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			return "", infraerrors.BadRequest("INVALID_INPUT", "session json cannot be parsed")
		}
		token, _ := payload["accessToken"].(string)
		token = strings.TrimSpace(token)
		if token == "" {
			return "", infraerrors.BadRequest("INVALID_INPUT", "accessToken is missing in session json")
		}
		return token, nil
	}
	if len(raw) < 20 {
		return "", infraerrors.BadRequest("INVALID_INPUT", "session token looks incomplete")
	}
	return raw, nil
}

func buildDeliveryHint(mode string, payload string, contact string) string {
	if mode == ShopFulfillmentSessionTopup {
		if len(payload) > 10 {
			return payload[:6] + "***" + payload[len(payload)-4:]
		}
		return "***"
	}
	if len(payload) > 40 {
		return payload[:40] + "..."
	}
	return payload
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// SubmitUserDeliveryInfo 登录用户为自己的商城订单补交交付资料（Session / 收货邮箱 / 租期）。
// 与游客链路共用同一套加密存储与脱敏逻辑，后台发货页体验完全一致。
func (s *ShopService) SubmitUserDeliveryInfo(ctx context.Context, userID int64, orderID int64, payload string, rentalDuration string) error {
	if s.encryptor == nil {
		return infraerrors.InternalServer("ENCRYPTOR_UNAVAILABLE", "delivery encryption is not configured")
	}
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return infraerrors.BadRequest("INVALID_INPUT", "delivery info is required")
	}
	if len(payload) > 20000 {
		return infraerrors.BadRequest("INVALID_INPUT", "delivery info is too long")
	}

	var status, mode string
	err := s.db.QueryRowContext(ctx, `
SELECT status, COALESCE(snapshot_fulfillment_mode, 'manual')
FROM shop_orders
WHERE tenant_id = 1 AND id = $1 AND user_id = $2`, orderID, userID).Scan(&status, &mode)
	if errors.Is(err, sql.ErrNoRows) {
		return infraerrors.NotFound("SHOP_ORDER_NOT_FOUND", "order not found")
	}
	if err != nil {
		return err
	}
	if status == "cancelled" || status == "refunded" || status == "failed" {
		return infraerrors.Conflict("SHOP_ORDER_CLOSED", "order is closed and cannot be updated")
	}

	normalized := payload
	if mode == ShopFulfillmentSessionTopup {
		token, err := extractSessionToken(payload)
		if err != nil {
			return err
		}
		normalized = token
	}
	encrypted, err := s.encryptor.Encrypt(normalized)
	if err != nil {
		return infraerrors.InternalServer("ENCRYPT_FAILED", "failed to secure delivery info")
	}
	_, err = s.db.ExecContext(ctx, `
UPDATE shop_orders
SET delivery_payload = $3, delivery_hint = $4, delivery_submitted_at = NOW(),
    rental_duration = COALESCE(NULLIF($5, ''), rental_duration), updated_at = NOW()
WHERE id = $1 AND user_id = $2`, orderID, userID, encrypted, buildDeliveryHint(mode, normalized, ""), strings.TrimSpace(rentalDuration))
	return err
}

// ClaimGuestOrder 登录用户认领自己在首页以免登录方式下的订单。
// 校验订单号 + 联系方式，且只认领仍挂在游客账号下的订单，避免误关联他人订单。
func (s *ShopService) ClaimGuestOrder(ctx context.Context, userID int64, orderNo string, contact string) error {
	orderNo = strings.TrimSpace(orderNo)
	contact = normalizeGuestContact(contact)
	if orderNo == "" {
		return infraerrors.BadRequest("INVALID_INPUT", "order number is required")
	}
	if err := validateGuestContact(contact); err != nil {
		return err
	}
	guestUserID, err := s.resolveGuestUserID(ctx)
	if err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `
UPDATE shop_orders
SET user_id = $4, updated_at = NOW()
WHERE tenant_id = 1 AND order_no = $1 AND guest_contact = $2 AND user_id = $3 AND status <> 'cancelled'`,
		orderNo, contact, guestUserID, userID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return infraerrors.NotFound("SHOP_ORDER_NOT_FOUND", "order not found, already claimed, or contact mismatch")
	}
	return nil
}

func generateGuestOrderNo(ctx context.Context, tx *sql.Tx) (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	for attempt := 0; attempt < 8; attempt++ {
		var sb strings.Builder
		// strings.Builder 的 Write* 永远返回 nil error，显式忽略以满足 errcheck。
		_, _ = sb.WriteString(guestOrderNoPrefix)
		_, _ = sb.WriteString(time.Now().Format("20060102"))
		for i := 0; i < guestOrderNoRandomLen; i++ {
			idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
			if err != nil {
				return "", err
			}
			_ = sb.WriteByte(alphabet[idx.Int64()])
		}
		candidate := sb.String()
		var exists int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM shop_orders WHERE order_no = $1`, candidate).Scan(&exists); err != nil {
			return "", err
		}
		if exists == 0 {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("failed to allocate unique order number")
}
