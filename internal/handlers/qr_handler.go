package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/winchester/shorturls/internal/repository"
	"github.com/winchester/shorturls/internal/utils"
)

type QRHandler struct {
	linkRepo      *repository.LinkRepo
	analyticsRepo *repository.AnalyticsRepo
	rdb           *redis.Client
}

func NewQRHandler(linkRepo *repository.LinkRepo, analyticsRepo *repository.AnalyticsRepo, rdb *redis.Client) *QRHandler {
	return &QRHandler{linkRepo: linkRepo, analyticsRepo: analyticsRepo, rdb: rdb}
}

func (h *QRHandler) GetQR(c *gin.Context) {
	alias := c.Param("alias")
	ctx := c.Request.Context()

	link, err := h.linkRepo.GetByAlias(ctx, alias)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if link == nil || link.IsDeleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}

	qrKey := fmt.Sprintf("qr:%s", alias)
	cached, err := h.rdb.Get(ctx, qrKey).Bytes()
	if err == nil && len(cached) > 0 {
		c.Data(http.StatusOK, "image/png", cached)
		go h.analyticsRepo.RecordQRScan(context.Background(), link.ID)
		return
	}

	pngData, err := utils.GenerateQRPNG(link.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate QR code"})
		return
	}

	h.rdb.Set(ctx, qrKey, pngData, 24*time.Hour)

	go h.analyticsRepo.RecordQRScan(context.Background(), link.ID)

	c.Data(http.StatusOK, "image/png", pngData)
}

func QRFromBytes(png []byte) bool {
	return len(png) > 0 && bytes.HasPrefix(png, []byte{0x89, 0x50, 0x4E, 0x47})
}
