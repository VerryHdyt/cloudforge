package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/VerryHdyt/cloudforge/backend/internal/services"
	"github.com/containerd/errdefs"
	"github.com/gin-gonic/gin"
)

type DockerHandler struct {
	Docker *services.DockerService
	Audit  *services.AuditService
}

type ContainerResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Image   string `json:"image"`
	State   string `json:"state"`
	Status  string `json:"status"`
	Created int64  `json:"created"`
}

func (h *DockerHandler) logAudit(c *gin.Context, action string, containerID string) {
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
		"containers",
		containerID,
	); err != nil {
		log.Printf("Failed to record Docker audit log: %v", err)
	}
}

func (h *DockerHandler) handleDockerError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrContainerIDRequired):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Container ID is required",
		})

	case errors.Is(err, services.ErrInvalidContainerID),
		errdefs.IsInvalidArgument(err):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid container ID",
		})

	case errors.Is(err, services.ErrProtectedContainer),
		errdefs.IsPermissionDenied(err):
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Operation is not allowed for this container",
		})

	case errdefs.IsNotFound(err):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Container not found",
		})

	case errdefs.IsConflict(err),
		errdefs.IsAlreadyExists(err),
		errdefs.IsFailedPrecondition(err):
		c.JSON(http.StatusConflict, gin.H{
			"error": "Container operation conflicts with its current state",
		})

	case errdefs.IsUnavailable(err),
		errdefs.IsDeadlineExceeded(err),
		errdefs.IsCanceled(err),
		errdefs.IsInternal(err),
		errdefs.IsUnknown(err):
		log.Printf("Docker Engine error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Docker Engine operation failed",
		})

	default:
		log.Printf("Unexpected Docker error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "An internal Docker operation error occurred",
		})
	}
}

func (h *DockerHandler) ListContainers(c *gin.Context) {
	result, err := h.Docker.ListContainers(c.Request.Context())
	if err != nil {
		log.Printf("Failed to list Docker containers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve Docker containers",
		})
		return
	}

	containers := make([]ContainerResponse, 0, len(result.Items))

	for _, container := range result.Items {
		name := ""

		if len(container.Names) > 0 {
			name = strings.TrimPrefix(container.Names[0], "/")
		}

		containers = append(containers, ContainerResponse{
			ID:      container.ID,
			Name:    name,
			Image:   container.Image,
			State:   string(container.State),
			Status:  container.Status,
			Created: container.Created,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  containers,
		"total": len(containers),
	})
}

func (h *DockerHandler) StartContainer(c *gin.Context) {
	id := c.Param("id")

	if err := h.Docker.StartContainer(c.Request.Context(), id); err != nil {
		h.handleDockerError(c, err)
		return
	}

	h.logAudit(c, "start", id)

	c.JSON(http.StatusOK, gin.H{
		"message":      "Container started successfully",
		"container_id": id,
	})
}

func (h *DockerHandler) StopContainer(c *gin.Context) {
	id := c.Param("id")

	if err := h.Docker.StopContainer(c.Request.Context(), id); err != nil {
		h.handleDockerError(c, err)
		return
	}

	h.logAudit(c, "stop", id)

	c.JSON(http.StatusOK, gin.H{
		"message":      "Container stopped successfully",
		"container_id": id,
	})
}

func (h *DockerHandler) RestartContainer(c *gin.Context) {
	id := c.Param("id")

	if err := h.Docker.RestartContainer(c.Request.Context(), id); err != nil {
		h.handleDockerError(c, err)
		return
	}

	h.logAudit(c, "restart", id)

	c.JSON(http.StatusOK, gin.H{
		"message":      "Container restarted successfully",
		"container_id": id,
	})
}

func (h *DockerHandler) GetContainerLogs(c *gin.Context) {
	id := c.Param("id")

	tail := c.DefaultQuery("tail", "100")

	tailCount, err := strconv.Atoi(tail)
	if err != nil || tailCount < 1 || tailCount > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tail must be a number between 1 and 1000",
		})
		return
	}

	timestamps := false

	if value, exists := c.GetQuery("timestamps"); exists {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Timestamps must be true or false",
			})
			return
		}

		timestamps = parsed
	}

	logs, err := h.Docker.GetContainerLogs(
		c.Request.Context(),
		id,
		strconv.Itoa(tailCount),
		timestamps,
	)
	if err != nil {
		h.handleDockerError(c, err)
		return
	}

	h.logAudit(c, "logs", id)

	c.JSON(http.StatusOK, gin.H{
		"container_id": id,
		"tail":         tailCount,
		"timestamps":   timestamps,
		"logs":         logs,
	})
}

// GetContainerStats retrieves CPU, memory, network and block I/O statistics.
func (h *DockerHandler) GetContainerStats(c *gin.Context) {
	id := c.Param("id")

	stats, err := h.Docker.GetContainerStats(
		c.Request.Context(),
		id,
	)
	if err != nil {
		h.handleDockerError(c, err)
		return
	}

	h.logAudit(c, "stats", id)

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}
