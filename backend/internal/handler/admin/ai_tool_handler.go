package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type AIToolHandler struct {
	toolService *service.AIToolService
}

func NewAIToolHandler(toolService *service.AIToolService) *AIToolHandler {
	return &AIToolHandler{toolService: toolService}
}

type aiToolRequest struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Status      string `json:"status"`
	SortOrder   int    `json:"sort_order"`
}

func (h *AIToolHandler) List(c *gin.Context) {
	items, err := h.toolService.AdminList(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

func (h *AIToolHandler) Create(c *gin.Context) {
	var req aiToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	item, err := h.toolService.Create(c.Request.Context(), req.toServiceInput())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *AIToolHandler) Update(c *gin.Context) {
	id, err := service.ParseAIToolID(c.Param("id"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	var req aiToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	item, err := h.toolService.Update(c.Request.Context(), id, req.toServiceInput())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *AIToolHandler) Delete(c *gin.Context) {
	id, err := service.ParseAIToolID(c.Param("id"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := h.toolService.Delete(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "tool archived"})
}

func (r aiToolRequest) toServiceInput() service.UpsertAIToolInput {
	return service.UpsertAIToolInput{
		Name:        r.Name,
		Category:    r.Category,
		Description: r.Description,
		URL:         r.URL,
		Status:      r.Status,
		SortOrder:   r.SortOrder,
	}
}
