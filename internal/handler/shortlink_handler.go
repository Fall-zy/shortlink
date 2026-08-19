package handler

import (
	"fmt"
	"net/http"
	"shortlink/config"
	"shortlink/internal/model"
	"shortlink/internal/service"
	"shortlink/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ShortLinkHandler struct {
	svc          *service.ShortLinkSvc
	accessLogSvc *service.AccessLogSvc
}

func NewShortLinkHandler(svc *service.ShortLinkSvc, accessLogSvc *service.AccessLogSvc) *ShortLinkHandler {
	return &ShortLinkHandler{svc: svc, accessLogSvc: accessLogSvc}
}

func (h *ShortLinkHandler) CreateShortLink(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required,url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Logger.Warn("创建短链接-参数无效",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	link, err := h.svc.CreateShortLink(req.URL)
	if err != nil {
		utils.Logger.Error("创建短链接-服务错误",
			zap.String("url", req.URL),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成失败"})
		return
	}
	//从配置读取域名
	baseURL := config.AppConfig.Server.BaseURL
	shortURL := fmt.Sprintf("%s/r/%s", baseURL, link.ShortCode)

	utils.Logger.Info("短链接创建成功",
		zap.String("short_code", link.ShortCode),
		zap.String("original_url", link.OriginalURL),
	)

	c.JSON(http.StatusCreated, gin.H{
		"short_url": shortURL,
		"code":      link.ShortCode,
	})
}

func (h *ShortLinkHandler) Redirect(c *gin.Context) {
	code := c.Param("code")
	originalURL, err := h.svc.GetOriginalURL(code)
	if err != nil {
		utils.Logger.Warn("短链接跳转-未找到",
			zap.String("code", code),
			zap.Error(err),
		)
		c.JSON(http.StatusNotFound, gin.H{"error": "短链接不存在或已失效"})
		return
	}

	logEntry := &model.AccessLog{
		ShortCode:  code,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
		Referer:    c.Request.Referer(),
		AccessTime: time.Now(),
	}
	h.accessLogSvc.AsyncLog(logEntry)

	utils.Logger.Info("短链接跳转",
		zap.String("code", code),
		zap.String("redirect_to", originalURL),
	)
	c.Redirect(http.StatusFound, originalURL)
}

func (h *ShortLinkHandler) GetStats(c *gin.Context) {
	code := c.Param("code")
	stats, err := h.accessLogSvc.GetStats(code)
	if err != nil {
		utils.Logger.Error("获取统计数据失败", zap.String("code", code), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计失败"})
		return
	}
	c.JSON(http.StatusOK, stats)
}
