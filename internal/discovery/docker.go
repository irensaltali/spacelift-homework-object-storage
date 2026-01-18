package discovery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// MinioInstance represents a discovered Minio instance.
type MinioInstance struct {
	ID        string
	Host      string
	Port      string
	AccessKey string
	SecretKey string
}

// DiscoverInstances finds all Minio instances in the Docker daemon.
func DiscoverInstances(ctx context.Context) ([]MinioInstance, error) {
	// Use docker inspect to get container details
	cmd := exec.CommandContext(ctx, "docker", "ps", "--filter", "name=amazin-object-storage-node", "--format", "{{.ID}}")

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	containerIDs := strings.Fields(out.String())
	if len(containerIDs) == 0 {
		return nil, fmt.Errorf("no minio instances found")
	}

	var instances []MinioInstance

	for _, id := range containerIDs {
		instance, err := extractInstanceInfo(ctx, id)
		if err != nil {
			fmt.Printf("warning: failed to extract info from container %s: %v\n", id[:12], err)
			continue
		}

		instances = append(instances, instance)
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("no valid minio instances found")
	}

	return instances, nil
}

// extractInstanceInfo extracts Minio connection details from a container.
func extractInstanceInfo(ctx context.Context, containerID string) (MinioInstance, error) {
	// Get container inspect output as JSON
	cmd := exec.CommandContext(ctx, "docker", "inspect", containerID)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return MinioInstance{}, fmt.Errorf("failed to inspect container: %w", err)
	}

	var inspectData []map[string]any
	if err := json.Unmarshal(out.Bytes(), &inspectData); err != nil {
		return MinioInstance{}, fmt.Errorf("failed to parse container inspect output: %w", err)
	}

	if len(inspectData) == 0 {
		return MinioInstance{}, fmt.Errorf("no container data returned")
	}

	container := inspectData[0]

	instance := MinioInstance{
		ID:   containerID[:12],
		Port: "9000",
	}

	// Extract IP address from NetworkSettings
	if networkSettings, ok := container["NetworkSettings"].(map[string]any); ok {
		if networks, ok := networkSettings["Networks"].(map[string]any); ok {
			for _, network := range networks {
				if netData, ok := network.(map[string]any); ok {
					if ip, ok := netData["IPAddress"].(string); ok && ip != "" {
						instance.Host = ip
						break
					}
				}
			}
		}
	}

	if instance.Host == "" {
		return MinioInstance{}, fmt.Errorf("no ip address found for container")
	}

	// Extract credentials from environment variables
	if config, ok := container["Config"].(map[string]any); ok {
		if env, ok := config["Env"].([]any); ok {
			for _, e := range env {
				if envStr, ok := e.(string); ok {
					if strings.HasPrefix(envStr, "MINIO_ACCESS_KEY=") {
						instance.AccessKey = strings.TrimPrefix(envStr, "MINIO_ACCESS_KEY=")
					}
					if strings.HasPrefix(envStr, "MINIO_SECRET_KEY=") {
						instance.SecretKey = strings.TrimPrefix(envStr, "MINIO_SECRET_KEY=")
					}
				}
			}
		}
	}

	if instance.AccessKey == "" || instance.SecretKey == "" {
		return MinioInstance{}, fmt.Errorf("missing minio credentials in environment")
	}

	return instance, nil
}
