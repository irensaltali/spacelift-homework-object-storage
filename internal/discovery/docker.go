package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strings"
	"time"
)

const (
	minioContainerNamePattern = "amazin-object-storage-node"
	minioAPIPort              = "9000"
	dockerSocketPath          = "/var/run/docker.sock"
	dockerAPIVersion          = "v1.41"
)

// MinioInstance represents a discovered Minio instance.
type MinioInstance struct {
	ID        string
	Host      string
	Port      string
	AccessKey string
	SecretKey string
}

type dockerClient interface {
	ListContainers(ctx context.Context, nameFilter string) ([]dockerContainerSummary, error)
	InspectContainer(ctx context.Context, containerID string) (dockerContainerInspect, error)
}

type dockerEngineClient struct {
	httpClient *http.Client
}

type dockerContainerSummary struct {
	ID    string   `json:"Id"`
	Names []string `json:"Names"`
}

type dockerContainerInspect struct {
	Config struct {
		Env []string `json:"Env"`
	} `json:"Config"`
	NetworkSettings struct {
		Networks map[string]struct {
			IPAddress string `json:"IPAddress"`
		} `json:"Networks"`
	} `json:"NetworkSettings"`
}

// DiscoverInstances finds all Minio instances in the Docker daemon.
func DiscoverInstances(ctx context.Context) ([]MinioInstance, error) {
	engineClient := newDockerEngineClient()
	return discoverInstances(ctx, engineClient)
}

func discoverInstances(ctx context.Context, dockerClient dockerClient) ([]MinioInstance, error) {
	containers, err := dockerClient.ListContainers(ctx, minioContainerNamePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}
	if len(containers) == 0 {
		return nil, fmt.Errorf("no minio instances found")
	}

	sort.Slice(containers, func(i, j int) bool {
		iName := primaryContainerName(containers[i].Names)
		jName := primaryContainerName(containers[j].Names)
		if iName == jName {
			return containers[i].ID < containers[j].ID
		}
		return iName < jName
	})

	instances := make([]MinioInstance, 0, len(containers))
	for _, listedContainer := range containers {
		inspectData, err := dockerClient.InspectContainer(ctx, listedContainer.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to inspect container %s: %w", shortContainerID(listedContainer.ID), err)
		}

		instance, err := extractInstanceInfo(listedContainer.ID, inspectData)
		if err != nil {
			return nil, fmt.Errorf("invalid minio container %s: %w", shortContainerID(listedContainer.ID), err)
		}

		instances = append(instances, instance)
	}

	return instances, nil
}

// extractInstanceInfo extracts Minio connection details from Docker inspect data.
func extractInstanceInfo(containerID string, inspectData dockerContainerInspect) (MinioInstance, error) {
	instance := MinioInstance{
		ID:   shortContainerID(containerID),
		Port: minioAPIPort,
	}

	networkNames := make([]string, 0, len(inspectData.NetworkSettings.Networks))
	for networkName := range inspectData.NetworkSettings.Networks {
		networkNames = append(networkNames, networkName)
	}
	sort.Strings(networkNames)

	for _, networkName := range networkNames {
		network := inspectData.NetworkSettings.Networks[networkName]
		if strings.TrimSpace(network.IPAddress) != "" {
			instance.Host = network.IPAddress
			break
		}
	}

	if instance.Host == "" {
		return MinioInstance{}, fmt.Errorf("no ip address found for container")
	}

	accessKey, secretKey := extractCredentials(inspectData.Config.Env)
	if accessKey == "" || secretKey == "" {
		return MinioInstance{}, fmt.Errorf("missing minio credentials in environment")
	}

	instance.AccessKey = accessKey
	instance.SecretKey = secretKey
	return instance, nil
}

func extractCredentials(envVars []string) (accessKey, secretKey string) {
	for _, envVar := range envVars {
		switch {
		case strings.HasPrefix(envVar, "MINIO_ACCESS_KEY="):
			accessKey = strings.TrimPrefix(envVar, "MINIO_ACCESS_KEY=")
		case strings.HasPrefix(envVar, "MINIO_SECRET_KEY="):
			secretKey = strings.TrimPrefix(envVar, "MINIO_SECRET_KEY=")
		}
	}

	// Support newer MinIO variable names as a fallback.
	if accessKey == "" {
		accessKey = readEnvValue(envVars, "MINIO_ROOT_USER")
	}
	if secretKey == "" {
		secretKey = readEnvValue(envVars, "MINIO_ROOT_PASSWORD")
	}

	return accessKey, secretKey
}

func readEnvValue(envVars []string, key string) string {
	prefix := key + "="
	for _, envVar := range envVars {
		if strings.HasPrefix(envVar, prefix) {
			return strings.TrimPrefix(envVar, prefix)
		}
	}

	return ""
}

func shortContainerID(containerID string) string {
	if len(containerID) <= 12 {
		return containerID
	}

	return containerID[:12]
}

func primaryContainerName(containerNames []string) string {
	if len(containerNames) == 0 {
		return ""
	}

	names := append([]string(nil), containerNames...)
	slices.Sort(names)
	return strings.TrimPrefix(names[0], "/")
}

func newDockerEngineClient() *dockerEngineClient {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dialer := net.Dialer{Timeout: 5 * time.Second}
			return dialer.DialContext(ctx, "unix", dockerSocketPath)
		},
	}

	return &dockerEngineClient{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   10 * time.Second,
		},
	}
}

func (dc *dockerEngineClient) ListContainers(ctx context.Context, nameFilter string) ([]dockerContainerSummary, error) {
	filters := map[string][]string{
		"name": {nameFilter},
	}
	filtersJSON, err := json.Marshal(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal container filters: %w", err)
	}

	path := fmt.Sprintf("/%s/containers/json?filters=%s", dockerAPIVersion, url.QueryEscape(string(filtersJSON)))

	var containers []dockerContainerSummary
	if err := dc.getJSON(ctx, path, &containers); err != nil {
		return nil, err
	}

	return containers, nil
}

func (dc *dockerEngineClient) InspectContainer(ctx context.Context, containerID string) (dockerContainerInspect, error) {
	path := fmt.Sprintf("/%s/containers/%s/json", dockerAPIVersion, containerID)

	var inspectData dockerContainerInspect
	if err := dc.getJSON(ctx, path, &inspectData); err != nil {
		return dockerContainerInspect{}, err
	}

	return inspectData, nil
}

func (dc *dockerEngineClient) getJSON(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://docker"+path, nil)
	if err != nil {
		return fmt.Errorf("failed to create docker request: %w", err)
	}

	resp, err := dc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("docker request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("docker API returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("failed to decode docker response: %w", err)
	}

	return nil
}
