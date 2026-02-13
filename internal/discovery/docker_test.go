package discovery

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeDockerClient struct {
	listResult    []dockerContainerSummary
	listErr       error
	inspectResult map[string]dockerContainerInspect
	inspectErr    error
}

func (f *fakeDockerClient) ListContainers(ctx context.Context, nameFilter string) ([]dockerContainerSummary, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listResult, nil
}

func (f *fakeDockerClient) InspectContainer(ctx context.Context, containerID string) (dockerContainerInspect, error) {
	if f.inspectErr != nil {
		return dockerContainerInspect{}, f.inspectErr
	}
	inspectData, ok := f.inspectResult[containerID]
	if !ok {
		return dockerContainerInspect{}, errors.New("inspect data not found")
	}
	return inspectData, nil
}

func TestDiscoverInstances_Sorted(t *testing.T) {
	client := &fakeDockerClient{
		listResult: []dockerContainerSummary{
			{ID: "bbbbbb222222333333", Names: []string{"/amazin-object-storage-node-2"}},
			{ID: "aaaaaa111111333333", Names: []string{"/amazin-object-storage-node-1"}},
		},
		inspectResult: map[string]dockerContainerInspect{
			"aaaaaa111111333333": newInspectData("172.17.0.2", "ring", "treepotato"),
			"bbbbbb222222333333": newInspectData("172.17.0.3", "maglev", "baconpapaya"),
		},
	}

	instances, err := discoverInstances(context.Background(), client)
	if err != nil {
		t.Fatalf("discoverInstances() unexpected error: %v", err)
	}

	if len(instances) != 2 {
		t.Fatalf("discoverInstances() returned %d instances, want 2", len(instances))
	}
	if instances[0].ID != "aaaaaa111111" {
		t.Fatalf("first instance ID = %s, want aaaaaa111111", instances[0].ID)
	}
	if instances[1].ID != "bbbbbb222222" {
		t.Fatalf("second instance ID = %s, want bbbbbb222222", instances[1].ID)
	}
}

func TestDiscoverInstances_NoContainers(t *testing.T) {
	client := &fakeDockerClient{}

	_, err := discoverInstances(context.Background(), client)
	if err == nil {
		t.Fatal("discoverInstances() expected error for empty list")
	}
	if !strings.Contains(err.Error(), "no minio instances found") {
		t.Fatalf("error = %v, want no minio instances found", err)
	}
}

func TestExtractInstanceInfo(t *testing.T) {
	inspect := newInspectData("172.17.0.2", "access", "secret")

	instance, err := extractInstanceInfo("1234567890abcdef", inspect)
	if err != nil {
		t.Fatalf("extractInstanceInfo() unexpected error: %v", err)
	}

	if instance.ID != "1234567890ab" {
		t.Fatalf("ID = %s, want 1234567890ab", instance.ID)
	}
	if instance.Host != "172.17.0.2" {
		t.Fatalf("Host = %s, want 172.17.0.2", instance.Host)
	}
	if instance.Port != "9000" {
		t.Fatalf("Port = %s, want 9000", instance.Port)
	}
	if instance.AccessKey != "access" || instance.SecretKey != "secret" {
		t.Fatalf("credentials mismatch: got %s/%s", instance.AccessKey, instance.SecretKey)
	}
}

func TestExtractInstanceInfo_MissingIP(t *testing.T) {
	inspect := dockerContainerInspect{}
	inspect.Config.Env = []string{"MINIO_ACCESS_KEY=access", "MINIO_SECRET_KEY=secret"}
	inspect.NetworkSettings.Networks = map[string]struct {
		IPAddress string `json:"IPAddress"`
	}{
		"bridge": {},
	}

	_, err := extractInstanceInfo("123", inspect)
	if err == nil {
		t.Fatal("extractInstanceInfo() expected error")
	}
	if !strings.Contains(err.Error(), "no ip address found") {
		t.Fatalf("error = %v, want no ip address found", err)
	}
}

func TestExtractCredentials_RootUserFallback(t *testing.T) {
	access, secret := extractCredentials([]string{
		"MINIO_ROOT_USER=root-user",
		"MINIO_ROOT_PASSWORD=root-password",
	})

	if access != "root-user" || secret != "root-password" {
		t.Fatalf("credentials = %s/%s, want root-user/root-password", access, secret)
	}
}

func newInspectData(ipAddress, accessKey, secretKey string) dockerContainerInspect {
	inspect := dockerContainerInspect{}
	inspect.Config.Env = []string{
		"MINIO_ACCESS_KEY=" + accessKey,
		"MINIO_SECRET_KEY=" + secretKey,
	}
	inspect.NetworkSettings.Networks = map[string]struct {
		IPAddress string `json:"IPAddress"`
	}{
		"amazin-object-storage": {
			IPAddress: ipAddress,
		},
	}
	return inspect
}
