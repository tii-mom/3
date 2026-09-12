package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

/**
 * 商城品类（迁移 208）。
 *
 * 迁移 207 把品类写成代码枚举 + CHECK 约束，导致后台无法自行新增品类。
 * 从 208 起品类集合由 shop_categories 表定义，后台可自由增删改；
 * 商品的 category 列存 slug，前端按 slug 分组展示。
 */

const (
	shopCategoryLabelMaxLen = 64
	shopCategoryBlurbMaxLen = 255
)

// ShopCategory 品类。
type ShopCategory struct {
	ID        int64  `json:"id"`
	Slug      string `json:"slug"`
	Label     string `json:"label"`
	Blurb     string `json:"blurb"`
	SortOrder int    `json:"sort_order"`
	Enabled   bool   `json:"enabled"`
	// ProductCount 仅后台列表返回，用于提示「该分类下有 N 个商品」并作为删除前置校验。
	ProductCount int    `json:"product_count"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

// UpsertShopCategoryInput 后台新建 / 更新品类的入参。
// Slug 仅在创建时生效；更新时不可修改（改了会孤立已归属的商品）。
type UpsertShopCategoryInput struct {
	Slug      string
	Label     string
	Blurb     string
	SortOrder int
	Enabled   bool
}

const shopCategoryColumns = `id, slug, label, blurb, sort_order, enabled, created_at, updated_at`

// ---------------------------------------------------------------------------
// 查询
// ---------------------------------------------------------------------------

// ListPublicCategories 官网首页用：只返回启用中的品类，按 sort_order 升序。
func (s *ShopService) ListPublicCategories(ctx context.Context) ([]ShopCategory, error) {
	return s.listCategories(ctx, false)
}

// AdminListCategories 后台用：返回全部品类（含停用），并带上各品类下的商品数。
func (s *ShopService) AdminListCategories(ctx context.Context) ([]ShopCategory, error) {
	return s.listCategories(ctx, true)
}

func (s *ShopService) listCategories(ctx context.Context, admin bool) ([]ShopCategory, error) {
	where := `tenant_id = 1`
	if !admin {
		where += ` AND enabled = TRUE`
	}

	// 商品数只统计未删除的商品（含未发布，便于后台判断「这个品类能不能删」）。
	query := `
SELECT c.id, c.slug, c.label, c.blurb, c.sort_order, c.enabled, c.created_at, c.updated_at`
	if admin {
		query += `, (SELECT COUNT(*) FROM shop_products p
                    WHERE p.tenant_id = c.tenant_id AND p.category = c.slug AND p.deleted_at IS NULL)`
	} else {
		query += `, 0`
	}
	query += `
FROM shop_categories c
WHERE c.` + where + `
ORDER BY c.sort_order ASC, c.id ASC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := []ShopCategory{}
	for rows.Next() {
		var item ShopCategory
		var created, updated time.Time
		if err := rows.Scan(&item.ID, &item.Slug, &item.Label, &item.Blurb, &item.SortOrder,
			&item.Enabled, &created, &updated, &item.ProductCount); err != nil {
			return nil, err
		}
		item.CreatedAt = created.Format(time.RFC3339)
		item.UpdatedAt = updated.Format(time.RFC3339)
		out = append(out, item)
	}
	return out, rows.Err()
}

// ensureCategoryExists 校验商品写入的品类是否已存在。
// 不检查 enabled：停用只是「不在官网展示」，不应阻止管理员给存量商品保留归属。
func (s *ShopService) ensureCategoryExists(ctx context.Context, slug string) error {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil
	}
	if !isValidCategorySlug(slug) {
		return infraerrors.BadRequest("INVALID_SHOP_CATEGORY", "品类标识格式不合法（小写字母开头，仅含小写字母/数字/下划线）")
	}
	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM shop_categories WHERE tenant_id = 1 AND slug = $1)`, slug).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return infraerrors.BadRequest("INVALID_SHOP_CATEGORY", "商品品类不存在，请先在商城后台创建该品类")
	}
	return nil
}

// ---------------------------------------------------------------------------
// 写操作
// ---------------------------------------------------------------------------

func (s *ShopService) CreateCategory(ctx context.Context, in UpsertShopCategoryInput) (*ShopCategory, error) {
	in = normalizeShopCategoryInput(in)
	if err := validateShopCategoryInput(in); err != nil {
		return nil, err
	}

	var item ShopCategory
	var created, updated time.Time
	err := s.db.QueryRowContext(ctx, `
INSERT INTO shop_categories (tenant_id, slug, label, blurb, sort_order, enabled)
VALUES (1, $1, $2, $3, $4, $5)
RETURNING `+shopCategoryColumns,
		in.Slug, in.Label, in.Blurb, in.SortOrder, in.Enabled).
		Scan(&item.ID, &item.Slug, &item.Label, &item.Blurb, &item.SortOrder,
			&item.Enabled, &created, &updated)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, infraerrors.Conflict("SHOP_CATEGORY_EXISTS", "该品类标识已存在，请换一个")
		}
		return nil, err
	}
	item.CreatedAt = created.Format(time.RFC3339)
	item.UpdatedAt = updated.Format(time.RFC3339)
	return &item, nil
}

// UpdateCategory 更新品类。Slug 不可变更：商品通过 slug 归属，改动会孤立存量商品。
func (s *ShopService) UpdateCategory(ctx context.Context, id int64, in UpsertShopCategoryInput) (*ShopCategory, error) {
	if id <= 0 {
		return nil, infraerrors.BadRequest("INVALID_ID", "invalid category id")
	}
	in = normalizeShopCategoryInput(in)
	if err := validateShopCategoryInput(in); err != nil {
		return nil, err
	}

	var item ShopCategory
	var created, updated time.Time
	err := s.db.QueryRowContext(ctx, `
UPDATE shop_categories
SET label = $2, blurb = $3, sort_order = $4, enabled = $5, updated_at = NOW()
WHERE tenant_id = 1 AND id = $1
RETURNING `+shopCategoryColumns,
		id, in.Label, in.Blurb, in.SortOrder, in.Enabled).
		Scan(&item.ID, &item.Slug, &item.Label, &item.Blurb, &item.SortOrder,
			&item.Enabled, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, infraerrors.NotFound("SHOP_CATEGORY_NOT_FOUND", "品类不存在")
	}
	if err != nil {
		return nil, err
	}
	item.CreatedAt = created.Format(time.RFC3339)
	item.UpdatedAt = updated.Format(time.RFC3339)
	return &item, nil
}

// DeleteCategory 删除品类。仍有商品归属该品类时拒绝删除，避免商品在官网分组中凭空消失。
func (s *ShopService) DeleteCategory(ctx context.Context, id int64) error {
	if id <= 0 {
		return infraerrors.BadRequest("INVALID_ID", "invalid category id")
	}

	var slug string
	err := s.db.QueryRowContext(ctx,
		`SELECT slug FROM shop_categories WHERE tenant_id = 1 AND id = $1`, id).Scan(&slug)
	if errors.Is(err, sql.ErrNoRows) {
		return infraerrors.NotFound("SHOP_CATEGORY_NOT_FOUND", "品类不存在")
	}
	if err != nil {
		return err
	}

	var inUse int64
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM shop_products WHERE tenant_id = 1 AND category = $1 AND deleted_at IS NULL`,
		slug).Scan(&inUse); err != nil {
		return err
	}
	if inUse > 0 {
		return infraerrors.Conflict("SHOP_CATEGORY_IN_USE",
			"该品类下还有商品，请先把这些商品改成其他品类再删除")
	}

	result, err := s.db.ExecContext(ctx, `DELETE FROM shop_categories WHERE tenant_id = 1 AND id = $1`, id)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return infraerrors.NotFound("SHOP_CATEGORY_NOT_FOUND", "品类不存在")
	}
	return nil
}

// ---------------------------------------------------------------------------
// 校验
// ---------------------------------------------------------------------------

func normalizeShopCategoryInput(in UpsertShopCategoryInput) UpsertShopCategoryInput {
	in.Slug = strings.ToLower(strings.TrimSpace(in.Slug))
	in.Label = strings.TrimSpace(in.Label)
	in.Blurb = strings.TrimSpace(in.Blurb)
	return in
}

func validateShopCategoryInput(in UpsertShopCategoryInput) error {
	if !isValidCategorySlug(in.Slug) {
		return infraerrors.BadRequest("INVALID_SHOP_CATEGORY",
			"品类标识需以小写字母开头，仅含小写字母、数字、下划线，最长 32 位")
	}
	if in.Label == "" {
		return infraerrors.BadRequest("INVALID_SHOP_CATEGORY", "品类名称不能为空")
	}
	if len([]rune(in.Label)) > shopCategoryLabelMaxLen {
		return infraerrors.BadRequest("INVALID_SHOP_CATEGORY", "品类名称过长")
	}
	if len([]rune(in.Blurb)) > shopCategoryBlurbMaxLen {
		return infraerrors.BadRequest("INVALID_SHOP_CATEGORY", "品类说明过长")
	}
	if in.SortOrder < 0 || in.SortOrder > 9999 {
		return infraerrors.BadRequest("INVALID_SHOP_CATEGORY", "排序值需在 0-9999 之间")
	}
	return nil
}

// isUniqueViolation 判断是否为唯一索引冲突（PostgreSQL 23505），
// 用于把「品类标识重复」转成可读的业务错误而不是 500。
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "23505") || strings.Contains(strings.ToLower(msg), "duplicate key")
}
