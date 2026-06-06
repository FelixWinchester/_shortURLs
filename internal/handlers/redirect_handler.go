package handlers

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/winchester/shorturls/internal/cache"
	"github.com/winchester/shorturls/internal/models"
	"github.com/winchester/shorturls/internal/repository"
	"github.com/winchester/shorturls/internal/utils"
)

type RedirectHandler struct {
	linkRepo      *repository.LinkRepo
	analyticsRepo *repository.AnalyticsRepo
	linkCache     *cache.LinkCache
}

func NewRedirectHandler(linkRepo *repository.LinkRepo, analyticsRepo *repository.AnalyticsRepo, linkCache *cache.LinkCache) *RedirectHandler {
	return &RedirectHandler{linkRepo: linkRepo, analyticsRepo: analyticsRepo, linkCache: linkCache}
}

func (h *RedirectHandler) Redirect(c *gin.Context) {
	alias := c.Param("alias")
	ctx := c.Request.Context()

	link, err := h.linkCache.Get(ctx, alias)
	if err != nil {
		log.Printf("cache get error: %v", err)
	}
	if link == nil {
		link, err = h.linkRepo.GetByAlias(ctx, alias)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		if link != nil {
			h.linkCache.Set(ctx, link)
		}
	}

	if link == nil {
		h.recordError(ctx, nil, c)
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}

	if link.IsDeleted {
		h.recordError(ctx, link, c)
		c.JSON(http.StatusGone, gin.H{"error": "link is deleted"})
		return
	}

	if link.IsDeactive {
		h.recordError(ctx, link, c)
		c.JSON(http.StatusGone, gin.H{"error": "link is deactivated"})
		return
	}

	if link.Lifetime != nil {
		expiresAt := link.CreatedAt.Add(time.Duration(*link.Lifetime) * time.Second)
		if time.Now().After(expiresAt) {
			h.recordError(ctx, link, c)
			c.JSON(http.StatusGone, gin.H{"error": "link has expired"})
			return
		}
	}

	if link.IsPrivate {
		token := c.Query("token")
		if token == "" || link.AccessToken == nil || !utils.VerifyToken(token, *link.AccessToken) {
			h.recordError(ctx, link, c)
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid or missing access token"})
			return
		}
	}

	if link.IsSingle && link.IsDeactive {
		h.recordError(ctx, link, c)
		c.JSON(http.StatusGone, gin.H{"error": "single-use link already used"})
		return
	}

	browser := parseBrowser(c.GetHeader("User-Agent"))

	if err := h.analyticsRepo.EnsureExists(ctx, link.ID); err != nil {
		log.Printf("ensure analytics error: %v", err)
	}

	if err := h.analyticsRepo.RecordSuccess(ctx, link.ID, browser); err != nil {
		log.Printf("record success error: %v", err)
	}

	if link.IsSingle {
		if err := h.analyticsRepo.DeactivateSingleUse(ctx, link.ID); err != nil {
			log.Printf("deactivate single error: %v", err)
		}
		link.IsDeactive = true
		h.linkCache.Set(ctx, link)
	}

	c.Redirect(http.StatusFound, link.URL)
}

func (h *RedirectHandler) recordError(ctx context.Context, link *models.Link, c *gin.Context) {
	if link != nil {
		browser := parseBrowser(c.GetHeader("User-Agent"))
		if err := h.analyticsRepo.EnsureExists(ctx, link.ID); err != nil {
			log.Printf("ensure analytics error: %v", err)
		}
		if err := h.analyticsRepo.RecordError(ctx, link.ID, browser); err != nil {
			log.Printf("record error: %v", err)
		}
	}
}

func parseBrowser(ua string) string {
	if ua == "" {
		return "Unknown"
	}
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "edg"):
		return "Edge"
	case strings.Contains(ua, "chrome"):
		return "Chrome"
	case strings.Contains(ua, "safari"):
		return "Safari"
	case strings.Contains(ua, "firefox"):
		return "Firefox"
	case strings.Contains(ua, "opera"):
		return "Opera"
	case strings.Contains(ua, "msie") || strings.Contains(ua, "trident"):
		return "Internet Explorer"
	default:
		return "Other"
	}
}
