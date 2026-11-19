package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	imageTypes "github.com/moby/moby/api/types/image"
	dockerclient "github.com/moby/moby/client"
)

var (
	ErrImageReferenceRequired = errors.New("image reference is required")
	ErrInvalidImageReference  = errors.New("image reference is invalid")
	ErrImageInUse             = errors.New("image is currently used by a container")
	ErrProtectedImage         = errors.New("image used by a protected container cannot be removed")
)

func validateImageReference(imageRef string) (string, error) {
	imageRef = strings.TrimSpace(imageRef)

	if imageRef == "" {
		return "", ErrImageReferenceRequired
	}

	if len(imageRef) > 255 ||
		strings.HasPrefix(imageRef, "-") ||
		strings.ContainsAny(imageRef, " \t\n\r") ||
		strings.Contains(imageRef, "://") {
		return "", ErrInvalidImageReference
	}

	return imageRef, nil
}

func (s *DockerService) ListImages(
	ctx context.Context,
) ([]imageTypes.Summary, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	result, err := s.Client.ImageList(
		ctx,
		dockerclient.ImageListOptions{
			All:        true,
			SharedSize: true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list Docker images: %w", err)
	}

	return result.Items, nil
}

func (s *DockerService) InspectImage(
	ctx context.Context,
	imageRef string,
) (dockerclient.ImageInspectResult, error) {
	imageRef, err := validateImageReference(imageRef)
	if err != nil {
		return dockerclient.ImageInspectResult{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	result, err := s.Client.ImageInspect(ctx, imageRef)
	if err != nil {
		return dockerclient.ImageInspectResult{},
			fmt.Errorf("failed to inspect Docker image: %w", err)
	}

	return result, nil
}

func (s *DockerService) PullImage(
	ctx context.Context,
	imageRef string,
) error {
	imageRef, err := validateImageReference(imageRef)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	result, err := s.Client.ImagePull(
		ctx,
		imageRef,
		dockerclient.ImagePullOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to pull Docker image: %w", err)
	}
	defer result.Close()

	if err := result.Wait(ctx); err != nil {
		return fmt.Errorf("Docker image pull failed: %w", err)
	}

	return nil
}

func (s *DockerService) RemoveImage(
	ctx context.Context,
	imageRef string,
) ([]imageTypes.DeleteResponse, error) {
	imageRef, err := validateImageReference(imageRef)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	imageInfo, err := s.Client.ImageInspect(ctx, imageRef)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect image before removal: %w", err)
	}

	containers, err := s.Client.ContainerList(
		ctx,
		dockerclient.ContainerListOptions{All: true},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check containers using image: %w", err)
	}

	for _, container := range containers.Items {
		if container.ImageID != imageInfo.ID {
			continue
		}

		for _, name := range container.Names {
			if strings.TrimPrefix(name, "/") == "cloudforge-postgres" {
				return nil, ErrProtectedImage
			}
		}

		return nil, ErrImageInUse
	}

	result, err := s.Client.ImageRemove(
		ctx,
		imageInfo.ID,
		dockerclient.ImageRemoveOptions{
			Force:         false,
			PruneChildren: false,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to remove Docker image: %w", err)
	}

	return result.Items, nil
}
