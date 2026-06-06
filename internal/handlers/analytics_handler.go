package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/winchester/shorturls/internal/models"
	"github.com/winchester/shorturls/internal/repository"
)

type AnalyticsHandler struct {
	linkRepo      *repository.LinkRepo
	analyticsRepo *repository.AnalyticsRepo
}

func NewAnalyticsHandler(linkRepo *repository.LinkRepo, analyticsRepo *repository.AnalyticsRepo) *AnalyticsHandler {
	return &AnalyticsHandler{linkRepo: linkRepo, analyticsRepo: analyticsRepo}
}

func (h *AnalyticsHandler) GetByAlias(c *gin.Context) {
	alias := c.Param("alias")
	ctx := c.Request.Context()

	link, err := h.linkRepo.GetByAlias(ctx, alias)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if link == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}

	a, err := h.analyticsRepo.GetByLinkID(ctx, link.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if a == nil {
		a = &models.Analytics{
			LinkID: link.ID,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"link": gin.H{
			"alias": link.Alias,
			"url":   link.URL,
		},
		"analytics": a,
	})
}

func (h *AnalyticsHandler) GetGlobal(c *gin.Context) {
	ctx := c.Request.Context()

	topLinks, err := h.analyticsRepo.GetTopLinks(ctx, 5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	totalSuccess, totalError, browserStats, err := h.analyticsRepo.GetTotals(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	result := models.GlobalAnalytics{
		TopLinks:     topLinks,
		TotalSuccess: totalSuccess,
		TotalError:   totalError,
		BrowserStats: browserStats,
	}

	c.JSON(http.StatusOK, result)
}
