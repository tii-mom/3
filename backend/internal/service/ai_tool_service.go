package service

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

type AIToolService struct {
	db *sql.DB
}

func NewAIToolService(db *sql.DB) *AIToolService {
	return &AIToolService{db: db}
}

type AITool struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Status      string `json:"status"`
	SortOrder   int    `json:"sort_order"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

type UpsertAIToolInput struct {
	Name        string
	Category    string
	Description string
	URL         string
	Status      string
	SortOrder   int
}

func (s *AIToolService) ListPublic(ctx context.Context) ([]AITool, error) {
	return s.list(ctx, false)
}

func (s *AIToolService) AdminList(ctx context.Context) ([]AITool, error) {
	return s.list(ctx, true)
}

func (s *AIToolService) list(ctx context.Context, admin bool) ([]AITool, error) {
	where := `tenant_id = 1 AND deleted_at IS NULL`
	if !admin {
		where += ` AND status = 'published'`
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, name, category, description, url, status, sort_order, created_at, updated_at
FROM ai_tools
WHERE `+where+`
ORDER BY sort_order ASC, id ASC`)
	if err != nil {
		if isAIToolsTableMissing(err) {
			return []AITool{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	items := []AITool{}
	for rows.Next() {
		item, err := scanAITool(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *AIToolService) Create(ctx context.Context, in UpsertAIToolInput) (*AITool, error) {
	in = normalizeAIToolInput(in)
	if err := validateAIToolInput(in); err != nil {
		return nil, err
	}
	row := s.db.QueryRowContext(ctx, `
INSERT INTO ai_tools (tenant_id, name, category, description, url, status, sort_order)
VALUES (1, $1, $2, $3, $4, $5, $6)
RETURNING id, name, category, description, url, status, sort_order, created_at, updated_at`,
		in.Name, in.Category, in.Description, in.URL, in.Status, in.SortOrder)
	item, err := scanAITool(row)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *AIToolService) Update(ctx context.Context, id int64, in UpsertAIToolInput) (*AITool, error) {
	if id <= 0 {
		return nil, infraerrors.BadRequest("INVALID_ID", "invalid tool id")
	}
	in = normalizeAIToolInput(in)
	if err := validateAIToolInput(in); err != nil {
		return nil, err
	}
	row := s.db.QueryRowContext(ctx, `
UPDATE ai_tools
SET name = $2, category = $3, description = $4, url = $5, status = $6, sort_order = $7, updated_at = NOW()
WHERE tenant_id = 1 AND id = $1 AND deleted_at IS NULL
RETURNING id, name, category, description, url, status, sort_order, created_at, updated_at`,
		id, in.Name, in.Category, in.Description, in.URL, in.Status, in.SortOrder)
	item, err := scanAITool(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, infraerrors.NotFound("AI_TOOL_NOT_FOUND", "tool not found")
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *AIToolService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return infraerrors.BadRequest("INVALID_ID", "invalid tool id")
	}
	result, err := s.db.ExecContext(ctx, `UPDATE ai_tools SET status = 'archived', deleted_at = NOW(), updated_at = NOW() WHERE tenant_id = 1 AND id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return infraerrors.NotFound("AI_TOOL_NOT_FOUND", "tool not found")
	}
	return nil
}

func normalizeAIToolInput(in UpsertAIToolInput) UpsertAIToolInput {
	in.Name = strings.TrimSpace(in.Name)
	in.Category = strings.TrimSpace(in.Category)
	if in.Category == "" {
		in.Category = "AI 编程"
	}
	in.Description = strings.TrimSpace(in.Description)
	in.URL = strings.TrimSpace(in.URL)
	in.Status = strings.TrimSpace(in.Status)
	if in.Status == "" {
		in.Status = "draft"
	}
	return in
}

func validateAIToolInput(in UpsertAIToolInput) error {
	if in.Name == "" {
		return infraerrors.BadRequest("INVALID_INPUT", "tool name is required")
	}
	if in.URL == "" {
		return infraerrors.BadRequest("INVALID_INPUT", "tool url is required")
	}
	parsed, err := url.Parse(in.URL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return infraerrors.BadRequest("INVALID_INPUT", "tool url must be a valid http or https url")
	}
	if in.Status != "draft" && in.Status != "published" && in.Status != "archived" {
		return infraerrors.BadRequest("INVALID_INPUT", "invalid tool status")
	}
	return nil
}

func isAIToolsTableMissing(err error) bool {
	return strings.Contains(err.Error(), `relation "ai_tools" does not exist`)
}

type aiToolScanner interface {
	Scan(dest ...any) error
}

func scanAITool(row aiToolScanner) (AITool, error) {
	var item AITool
	var created, updated time.Time
	err := row.Scan(&item.ID, &item.Name, &item.Category, &item.Description, &item.URL, &item.Status, &item.SortOrder, &created, &updated)
	if err != nil {
		return item, err
	}
	item.CreatedAt = created.Format(time.RFC3339)
	item.UpdatedAt = updated.Format(time.RFC3339)
	return item, nil
}

func ParseAIToolID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, infraerrors.BadRequest("INVALID_ID", "invalid tool id")
	}
	return id, nil
}
