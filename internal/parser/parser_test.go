package parser

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestParseComposeContent(t *testing.T) {
	content := []byte(`
x-fnpack:
  manifest:
    appname: "docker-chromium"
    version: "1.2.3"
    display_name: "浏览器"
    arch: "noarch"

services:
  chromium:
    image: registry.cn-guangzhou.aliyuncs.com/fnapp/trim-chromium:latest
    container_name: chromium
    ports:
      - 3000:3000
      - 3001:3001
`)

	compose, err := ParseComposeContent(content)
	if err != nil {
		t.Fatalf("ParseComposeContent failed: %v", err)
	}

	// Check manifest
	if compose.XFnpack.Manifest["appname"] != "docker-chromium" {
		t.Errorf("Expected appname 'docker-chromium', got %v", compose.XFnpack.Manifest["appname"])
	}

	if compose.XFnpack.Manifest["version"] != "1.2.3" {
		t.Errorf("Expected version '1.2.3', got %v", compose.XFnpack.Manifest["version"])
	}

	// Check services
	if len(compose.Services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(compose.Services))
	}

	chromium, ok := compose.Services["chromium"]
	if !ok {
		t.Fatal("Expected 'chromium' service")
	}

	if chromium.ContainerName != "chromium" {
		t.Errorf("Expected container_name 'chromium', got %s", chromium.ContainerName)
	}

	if len(chromium.Ports) != 2 {
		t.Errorf("Expected 2 ports, got %d", len(chromium.Ports))
	}
}

func TestExtractVariables(t *testing.T) {
	compose := &ComposeFile{
		Services: map[string]Service{
			"myapp": {
				ContainerName: "my-container",
				Ports:         []string{"8080:80", "443:443"},
			},
		},
	}

	vars := ExtractVariables(compose)

	if vars.ServiceName != "myapp" {
		t.Errorf("Expected ServiceName 'myapp', got %s", vars.ServiceName)
	}

	if vars.ContainerName != "my-container" {
		t.Errorf("Expected ContainerName 'my-container', got %s", vars.ContainerName)
	}

	if vars.FirstPort != "8080" {
		t.Errorf("Expected FirstPort '8080', got %s", vars.FirstPort)
	}
}

func TestExtractVariables_NoContainerName(t *testing.T) {
	compose := &ComposeFile{
		Services: map[string]Service{
			"webapp": {
				Ports: []string{"3000:3000"},
			},
		},
	}

	vars := ExtractVariables(compose)

	if vars.ServiceName != "webapp" {
		t.Errorf("Expected ServiceName 'webapp', got %s", vars.ServiceName)
	}

	// Should fall back to service name
	if vars.ContainerName != "webapp" {
		t.Errorf("Expected ContainerName 'webapp', got %s", vars.ContainerName)
	}
}

func TestExtractHostPort(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"3000", "3000"},
		{"3000:8080", "3000"},
		{"0.0.0.0:3000:8080", "3000"},
		{"3000:8080/tcp", "3000"},
		{"127.0.0.1:9000:9000/udp", "9000"},
	}

	for _, tt := range tests {
		result := extractHostPort(tt.input)
		if result != tt.expected {
			t.Errorf("extractHostPort(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestCleanComposeContent(t *testing.T) {
	content := []byte(`
x-fnpack:
  manifest:
    appname: test
services:
  app:
    image: nginx
`)

	cleaned, err := CleanComposeContent(content)
	if err != nil {
		t.Fatalf("CleanComposeContent failed: %v", err)
	}

	// Parse cleaned content to verify x-fnpack is removed
	var result map[string]interface{}
	if err := parseYAML(cleaned, &result); err != nil {
		t.Fatalf("Failed to parse cleaned content: %v", err)
	}

	if _, ok := result["x-fnpack"]; ok {
		t.Error("x-fnpack should be removed from cleaned content")
	}

	if _, ok := result["services"]; !ok {
		t.Error("services should be preserved in cleaned content")
	}
}

func TestGetManifestValue(t *testing.T) {
	manifest := map[string]interface{}{
		"appname": "myapp",
		"version": "1.0.0",
		"beta":    true,
		"count":   42,
	}

	tests := []struct {
		key      string
		defVal   string
		expected string
	}{
		{"appname", "default", "myapp"},
		{"version", "0.0.0", "1.0.0"},
		{"beta", "no", "yes"},
		{"count", "0", "42"},
		{"missing", "fallback", "fallback"},
	}

	for _, tt := range tests {
		result := GetManifestValue(manifest, tt.key, tt.defVal)
		if result != tt.expected {
			t.Errorf("GetManifestValue(%q, %q) = %q, expected %q", tt.key, tt.defVal, result, tt.expected)
		}
	}
}

func TestGetManifestValue_NilManifest(t *testing.T) {
	result := GetManifestValue(nil, "key", "default")
	if result != "default" {
		t.Errorf("Expected 'default' for nil manifest, got %q", result)
	}
}

// Helper function for testing
func parseYAML(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}


func TestExtractImageInfo(t *testing.T) {
	tests := []struct {
		image       string
		expectedOrg string
		expectedName string
	}{
		// Simple image name
		{"alpine", "alpine", "alpine"},
		{"alpine:latest", "alpine", "alpine"},
		{"nginx:1.21", "nginx", "nginx"},
		
		// org/image format
		{"lobehub/lobe-chat", "lobehub", "lobe-chat"},
		{"lobehub/lobe-chat:latest", "lobehub", "lobe-chat"},
		{"linuxserver/chromium:v1.0", "linuxserver", "chromium"},
		
		// registry/org/image format
		{"registry.cn-guangzhou.aliyuncs.com/fnapp/trim-chromium:latest", "fnapp", "trim-chromium"},
		{"ghcr.io/lobehub/lobe-chat:main", "lobehub", "lobe-chat"},
		{"docker.io/library/nginx:alpine", "library", "nginx"},
		
		// Empty
		{"", "", ""},
	}

	for _, tt := range tests {
		org, name := extractImageInfo(tt.image)
		if org != tt.expectedOrg {
			t.Errorf("extractImageInfo(%q) org = %q, expected %q", tt.image, org, tt.expectedOrg)
		}
		if name != tt.expectedName {
			t.Errorf("extractImageInfo(%q) name = %q, expected %q", tt.image, name, tt.expectedName)
		}
	}
}

func TestExtractVariables_WithImage(t *testing.T) {
	compose := &ComposeFile{
		Services: map[string]Service{
			"lobe-chat": {
				Image:         "lobehub/lobe-chat:latest",
				ContainerName: "lobe-chat",
				Ports:         []string{"3210:3210"},
			},
		},
	}

	vars := ExtractVariables(compose)

	if vars.ImageOrg != "lobehub" {
		t.Errorf("Expected ImageOrg 'lobehub', got %s", vars.ImageOrg)
	}

	if vars.ImageName != "lobe-chat" {
		t.Errorf("Expected ImageName 'lobe-chat', got %s", vars.ImageName)
	}
}
