package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/moby/moby/api/pkg/stdcopy"
	containerTypes "github.com/moby/moby/api/types/container"
	dockerclient "github.com/moby/moby/client"
)

var (
	ErrContainerIDRequired = errors.New("container ID is required")
	ErrInvalidContainerID  = errors.New("container ID is invalid")
	ErrProtectedContainer  = errors.New("operation on this protected container is not allowed")
)

var containerIDPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,127}$`)

type DockerService struct {
	Client *dockerclient.Client
}

type ContainerLogs struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

type ContainerStats struct {
	ContainerID   string  `json:"container_id"`
	Name          string  `json:"name"`
	Read          string  `json:"read"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryUsage   uint64  `json:"memory_usage"`
	MemoryLimit   uint64  `json:"memory_limit"`
	MemoryPercent float64 `json:"memory_percent"`
	NetworkRx     uint64  `json:"network_rx"`
	NetworkTx     uint64  `json:"network_tx"`
	BlockRead     uint64  `json:"block_read"`
	BlockWrite    uint64  `json:"block_write"`
}

func NewDockerService() (*DockerService, error) {
	client, err := dockerclient.New(dockerclient.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Docker client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.Ping(ctx, dockerclient.PingOptions{})
	if err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to connect to Docker Engine: %w", err)
	}

	return &DockerService{
		Client: client,
	}, nil
}

func validateContainerID(containerID string) (string, error) {
	containerID = strings.TrimSpace(containerID)

	if containerID == "" {
		return "", ErrContainerIDRequired
	}

	if !containerIDPattern.MatchString(containerID) {
		return "", ErrInvalidContainerID
	}

	return containerID, nil
}

func (s *DockerService) ListContainers(ctx context.Context) (dockerclient.ContainerListResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result, err := s.Client.ContainerList(ctx, dockerclient.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return dockerclient.ContainerListResult{}, fmt.Errorf("failed to list Docker containers: %w", err)
	}

	return result, nil
}

func (s *DockerService) StartContainer(ctx context.Context, containerID string) error {
	containerID, err := validateContainerID(containerID)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err = s.Client.ContainerStart(
		ctx,
		containerID,
		dockerclient.ContainerStartOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	return nil
}

func (s *DockerService) StopContainer(ctx context.Context, containerID string) error {
	containerID, err := validateContainerID(containerID)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	inspect, err := s.Client.ContainerInspect(
		ctx,
		containerID,
		dockerclient.ContainerInspectOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to inspect container: %w", err)
	}

	containerName := strings.TrimPrefix(inspect.Container.Name, "/")

	if containerName == "cloudforge-postgres" {
		return ErrProtectedContainer
	}

	if inspect.Container.Config != nil {
		labels := inspect.Container.Config.Labels

		if labels["com.docker.compose.service"] == "postgres" {
			return ErrProtectedContainer
		}
	}

	timeout := 10

	_, err = s.Client.ContainerStop(
		ctx,
		containerID,
		dockerclient.ContainerStopOptions{
			Timeout: &timeout,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	return nil
}

func (s *DockerService) RestartContainer(ctx context.Context, containerID string) error {
	containerID, err := validateContainerID(containerID)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	inspect, err := s.Client.ContainerInspect(
		ctx,
		containerID,
		dockerclient.ContainerInspectOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to inspect container: %w", err)
	}

	containerName := strings.TrimPrefix(inspect.Container.Name, "/")

	if containerName == "cloudforge-postgres" {
		return ErrProtectedContainer
	}

	if inspect.Container.Config != nil {
		labels := inspect.Container.Config.Labels

		if labels["com.docker.compose.service"] == "postgres" {
			return ErrProtectedContainer
		}
	}

	timeout := 10

	_, err = s.Client.ContainerRestart(
		ctx,
		containerID,
		dockerclient.ContainerRestartOptions{
			Timeout: &timeout,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to restart container: %w", err)
	}

	return nil
}

func (s *DockerService) GetContainerLogs(
	ctx context.Context,
	containerID string,
	tail string,
	timestamps bool,
) (*ContainerLogs, error) {
	containerID, err := validateContainerID(containerID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if tail == "" {
		tail = "100"
	}

	result, err := s.Client.ContainerLogs(
		ctx,
		containerID,
		dockerclient.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Tail:       tail,
			Timestamps: timestamps,
			Follow:     false,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve container logs: %w", err)
	}
	defer result.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	_, err = stdcopy.StdCopy(&stdout, &stderr, result)
	if err != nil {
		return nil, fmt.Errorf("failed to read container logs: %w", err)
	}

	return &ContainerLogs{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}, nil
}

func (s *DockerService) GetContainerStats(
	ctx context.Context,
	containerID string,
) (*ContainerStats, error) {
	containerID, err := validateContainerID(containerID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	result, err := s.Client.ContainerStats(
		ctx,
		containerID,
		dockerclient.ContainerStatsOptions{
			Stream:                false,
			IncludePreviousSample: true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve container stats: %w", err)
	}
	defer result.Body.Close()

	var stats containerTypes.StatsResponse

	if err := json.NewDecoder(result.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode container stats: %w", err)
	}

	return calculateContainerStats(stats), nil
}

func calculateContainerStats(stats containerTypes.StatsResponse) *ContainerStats {
	response := &ContainerStats{
		ContainerID: stats.ID,
		Name:        strings.TrimPrefix(stats.Name, "/"),
		Read:        stats.Read.Format(time.RFC3339),
	}

	// CPU usage percentage.
	cpuTotal := stats.CPUStats.CPUUsage.TotalUsage
	preCPU := stats.PreCPUStats.CPUUsage.TotalUsage

	systemCPU := stats.CPUStats.SystemUsage
	preSystemCPU := stats.PreCPUStats.SystemUsage

	if systemCPU > preSystemCPU && cpuTotal >= preCPU {
		cpuDelta := cpuTotal - preCPU
		systemDelta := systemCPU - preSystemCPU

		onlineCPUs := stats.CPUStats.OnlineCPUs

		if onlineCPUs == 0 {
			onlineCPUs = uint32(len(stats.CPUStats.CPUUsage.PercpuUsage))
		}

		if onlineCPUs > 0 {
			response.CPUPercent = float64(cpuDelta) /
				float64(systemDelta) *
				float64(onlineCPUs) * 100
		}
	}

	// Memory usage.
	response.MemoryUsage = stats.MemoryStats.Usage
	response.MemoryLimit = stats.MemoryStats.Limit

	if response.MemoryLimit > 0 {
		response.MemoryPercent = float64(response.MemoryUsage) /
			float64(response.MemoryLimit) * 100
	}

	// Network RX and TX.
	for _, network := range stats.Networks {
		response.NetworkRx += network.RxBytes
		response.NetworkTx += network.TxBytes
	}

	// Block I/O read and write.
	for _, entry := range stats.BlkioStats.IoServiceBytesRecursive {
		switch strings.ToLower(entry.Op) {
		case "read":
			response.BlockRead += entry.Value
		case "write":
			response.BlockWrite += entry.Value
		}
	}

	return response
}
