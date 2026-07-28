package handler

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

func (h *AIToolHandler) List(c *gin.Context) {
	items, err := h.toolService.ListPublic(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}
