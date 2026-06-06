package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/winchester/shorturls/internal/cache"
	"github.com/winchester/shorturls/internal/models"
	"github.com/winchester/shorturls/internal/repository"
	"github.com/winchester/shorturls/internal/utils"
)

type LinkHandler struct {
	repo  *repository.LinkRepo
	cache *cache.LinkCache
}

func NewLinkHandler(repo *repository.LinkRepo, cache *cache.LinkCache) *LinkHandler {
	return &LinkHandler{repo: repo, cache: cache}
}

func (h *LinkHandler) Create(c *gin.Context) {
	var req models.CreateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	alias := req.Alias
	if alias == "" {
		alias = utils.GenerateAlias()
	}

	taken, err := h.repo.IsAliasTaken(c.Request.Context(), alias)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if taken {
		c.JSON(http.StatusConflict, gin.H{"error": "alias already exists"})
		return
	}

	var accessToken string
	var accessTokenHash *string
	if req.IsPrivate {
		token, err := utils.GenerateToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		accessToken = token
		hash, err := utils.HashToken(token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		accessTokenHash = &hash
	}

	link := &models.Link{
		Alias:       alias,
		URL:         req.URL,
		Lifetime:    req.Lifetime,
		IsPrivate:   req.IsPrivate,
		IsSingle:    req.IsSingle,
		AccessToken: accessTokenHash,
	}

	if err := h.repo.Create(c.Request.Context(), link); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create link"})
		return
	}

	h.cache.Set(c.Request.Context(), link)

	link.AccessToken = nil
	resp := models.CreateLinkResponse{Link: link}
	if req.IsPrivate {
		resp.AccessToken = accessToken
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *LinkHandler) Get(c *gin.Context) {
	alias := c.Param("alias")

	link, err := h.repo.GetByAlias(c.Request.Context(), alias)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if link == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}

	link.AccessToken = nil
	c.JSON(http.StatusOK, link)
}

func (h *LinkHandler) Update(c *gin.Context) {
	alias := c.Param("alias")

	var req models.UpdateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, err := h.repo.GetByAlias(c.Request.Context(), alias)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}

	if req.Alias != nil && *req.Alias != alias {
		taken, err := h.repo.IsAliasTaken(c.Request.Context(), *req.Alias)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		if taken {
			c.JSON(http.StatusConflict, gin.H{"error": "new alias already exists"})
			return
		}
	}

	link := &models.Link{ID: existing.ID}
	if req.URL != nil {
		link.URL = *req.URL
	}
	if req.Alias != nil {
		link.Alias = *req.Alias
	}
	if req.Lifetime != nil {
		link.Lifetime = req.Lifetime
	}
	if req.IsPrivate != nil {
		link.IsPrivate = *req.IsPrivate
	}
	if req.IsSingle != nil {
		link.IsSingle = *req.IsSingle
	}
	if req.IsDeactive != nil {
		link.IsDeactive = *req.IsDeactive
	}

	if err := h.repo.Update(c.Request.Context(), link); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update link"})
		return
	}

	updated, err := h.repo.GetByID(c.Request.Context(), existing.ID)
	if err != nil || updated == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated link"})
		return
	}

	h.cache.Delete(c.Request.Context(), alias)
	if req.Alias != nil {
		h.cache.Delete(c.Request.Context(), *req.Alias)
	}
	h.cache.Set(c.Request.Context(), updated)

	updated.AccessToken = nil
	c.JSON(http.StatusOK, updated)
}

func (h *LinkHandler) Delete(c *gin.Context) {
	alias := c.Param("alias")

	if err := h.repo.SoftDelete(c.Request.Context(), alias); err != nil {
		if errors.Is(err, errors.New("link not found")) || err.Error() == "link not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	h.cache.Delete(c.Request.Context(), alias)
	c.JSON(http.StatusOK, gin.H{"message": "link deleted"})
}
