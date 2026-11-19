package handlers

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/VerryHdyt/cloudforge/backend/internal/services"
	"github.com/containerd/errdefs"
	"github.com/gin-gonic/gin"
)

// ListImages retrieves all Docker images.
func (h *DockerHandler) ListImages(c *gin.Context) {
	images, err := h.Docker.ListImages(c.Request.Context())
	if err != nil {
		log.Printf("Failed to list Docker images: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve Docker images",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  images,
		"total": len(images),
	})
}

// InspectImage retrieves detailed information about an image.
func (h *DockerHandler) InspectImage(c *gin.Context) {
	imageRef := c.Param("id")

	image, err := h.Docker.InspectImage(
		c.Request.Context(),
		imageRef,
	)
	if err != nil {
		h.handleImageError(c, err)
		return
	}

	h.logImageAudit(c, "inspect", imageRef)

	c.JSON(http.StatusOK, gin.H{
		"data": image,
	})
}

// PullImage downloads an image from a registry.
func (h *DockerHandler) PullImage(c *gin.Context) {
	var request struct {
		Image string `json:"image" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Image reference is required",
		})
		return
	}

	imageRef := strings.TrimSpace(request.Image)

	if err := h.Docker.PullImage(
		c.Request.Context(),
		imageRef,
	); err != nil {
		h.handleImageError(c, err)
		return
	}

	h.logImageAudit(c, "pull", imageRef)

	c.JSON(http.StatusOK, gin.H{
		"message": "Docker image pulled successfully",
		"image":   imageRef,
	})
}

// RemoveImage removes an image when it is not used by any container.
func (h *DockerHandler) RemoveImage(c *gin.Context) {
	imageRef := c.Param("id")

	deleted, err := h.Docker.RemoveImage(
		c.Request.Context(),
		imageRef,
	)
	if err != nil {
		h.handleImageError(c, err)
		return
	}

	h.logImageAudit(c, "remove", imageRef)

	c.JSON(http.StatusOK, gin.H{
		"message": "Docker image removed successfully",
		"image":   imageRef,
		"deleted": deleted,
	})
}

func (h *DockerHandler) logImageAudit(
	c *gin.Context,
	action string,
	imageRef string,
) {
	var userID *uint

	if value, exists := c.Get("user_id"); exists {
		if id, ok := value.(uint); ok {
			userID = &id
		}
	}

	if h.Audit == nil {
		log.Println("Docker audit service is not configured")
		return
	}

	if err := h.Audit.LogFromContext(
		c,
		userID,
		action,
		"images",
		imageRef,
	); err != nil {
		log.Printf("Failed to record Docker image audit log: %v", err)
	}
}

func (h *DockerHandler) handleImageError(
	c *gin.Context,
	err error,
) {
	switch {
	case errors.Is(err, services.ErrImageReferenceRequired):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Image reference is required",
		})

	case errors.Is(err, services.ErrInvalidImageReference),
		errdefs.IsInvalidArgument(err):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid image reference",
		})

	case errors.Is(err, services.ErrProtectedImage):
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Image is used by a protected container",
		})

	case errors.Is(err, services.ErrImageInUse),
		errdefs.IsConflict(err),
		errdefs.IsFailedPrecondition(err):
		c.JSON(http.StatusConflict, gin.H{
			"error": "Image is currently used by a container",
		})

	case errdefs.IsNotFound(err):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Docker image not found",
		})

	default:
		log.Printf("Docker image operation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Docker image operation failed",
		})
	}
}
