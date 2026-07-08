package handler

import (
	"fmt"
	"net/http"
	"shortlink/config"
	"shortlink/internal/service"

	"github.com/gin-gonic/gin"
)

type ShortLinkHandler struct {
	svc *service.ShortLinkSvc
}

func NewShortLinkHandler(svc *service.ShortLinkSvc) *ShortLinkHandler {
	return &ShortLinkHandler{svc: svc}
}

func (h *ShortLinkHandler) CreateShortLink(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required,url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	link, err := h.svc.CreateShortLink(req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	//从配置读取域名
	baseURL := config.AppConfig.Server.BaseURL
	shortURL := fmt.Sprintf("%s/r/%s", baseURL, link.ShortCode)
	c.JSON(http.StatusCreated, gin.H{
		"short_url": shortURL,
		"code":      link.ShortCode,
	})
}

func (h *ShortLinkHandler) Redirect(c *gin.Context) {
	code := c.Param("code")
	originalURL, err := h.svc.GetOriginalURL(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "短链接不存在或已失效"})
		return
	}
	c.Redirect(http.StatusMovedPermanently, originalURL)
}
