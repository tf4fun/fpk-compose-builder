package parser

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseComposeFile parses a docker-compose.yaml file and extracts x-fnpack and services
func ParseComposeFile(filePath string) (*ComposeFile, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read compose file: %w", err)
	}

	return ParseComposeContent(data)
}

// ParseComposeContent parses docker-compose content from bytes
func ParseComposeContent(data []byte) (*ComposeFile, error) {
	var compose ComposeFile
	if err := yaml.Unmarshal(data, &compose); err != nil {
		return nil, fmt.Errorf("failed to parse compose yaml: %w", err)
	}

	// Also parse raw content to extract custom files from x-fnpack
	var rawContent map[string]interface{}
	if err := yaml.Unmarshal(data, &rawContent); err != nil {
		return nil, fmt.Errorf("failed to parse raw yaml: %w", err)
	}

	// Extract raw x-fnpack content for custom file handling
	if xfnpack, ok := rawContent["x-fnpack"].(map[string]interface{}); ok {
		compose.XFnpack.RawContent = xfnpack
		compose.XFnpack.Files = extractCustomFiles(xfnpack)
	}

	return &compose, nil
}

// extractCustomFiles extracts all file paths and contents from x-fnpack
// All keys except "manifest" are treated as file paths with multi-line text content
func extractCustomFiles(xfnpack map[string]interface{}) map[string]string {
	files := make(map[string]string)

	for key, value := range xfnpack {
		// Skip manifest - it's handled separately as YAML object -> key=value
		if key == "manifest" {
			continue
		}

		// All other keys are file paths with string content
		if strValue, ok := value.(string); ok {
			files[key] = strValue
		}
	}

	return files
}

// ExtractVariables extracts template variables from the first service
func ExtractVariables(compose *ComposeFile) Variables {
	var vars Variables

	if len(compose.Services) == 0 {
		return vars
	}

	// Get the first service (sorted by name for consistency)
	serviceNames := make([]string, 0, len(compose.Services))
	for name := range compose.Services {
		serviceNames = append(serviceNames, name)
	}
	sort.Strings(serviceNames)

	firstServiceName := serviceNames[0]
	firstService := compose.Services[firstServiceName]

	vars.ServiceName = firstServiceName

	// Use container_name if specified, otherwise use service name
	if firstService.ContainerName != "" {
		vars.ContainerName = firstService.ContainerName
	} else {
		vars.ContainerName = firstServiceName
	}

	// Extract first port (host port from "host:container" format)
	if len(firstService.Ports) > 0 {
		vars.FirstPort = extractHostPort(firstService.Ports[0])
	}

	// Extract image organization and name
	vars.ImageOrg, vars.ImageName = extractImageInfo(firstService.Image)

	return vars
}

// extractImageInfo extracts organization and image name from docker image string
// Examples:
//   - "lobehub/lobe-chat:latest" -> org="lobehub", name="lobe-chat"
//   - "alpine:latest" -> org="alpine", name="alpine"
//   - "registry.example.com/org/image:tag" -> org="org", name="image"
func extractImageInfo(image string) (org, name string) {
	if image == "" {
		return "", ""
	}

	// Remove tag (everything after ":")
	if idx := strings.LastIndex(image, ":"); idx != -1 {
		// Make sure it's not part of a port (registry:port/image)
		if !strings.Contains(image[idx:], "/") {
			image = image[:idx]
		}
	}

	// Split by "/"
	parts := strings.Split(image, "/")

	switch len(parts) {
	case 1:
		// Simple image name: "alpine"
		return parts[0], parts[0]
	case 2:
		// org/image: "lobehub/lobe-chat"
		return parts[0], parts[1]
	default:
		// registry/org/image or deeper: "registry.example.com/org/image"
		// Return the second-to-last as org, last as name
		return parts[len(parts)-2], parts[len(parts)-1]
	}
}

// extractHostPort extracts the host port from a port mapping string
// Supports formats: "3000", "3000:8080", "0.0.0.0:3000:8080"
func extractHostPort(portMapping string) string {
	// Remove any protocol suffix (e.g., "/tcp", "/udp")
	portMapping = strings.Split(portMapping, "/")[0]

	parts := strings.Split(portMapping, ":")
	switch len(parts) {
	case 1:
		// Just port number: "3000"
		return parts[0]
	case 2:
		// host:container: "3000:8080"
		return parts[0]
	case 3:
		// ip:host:container: "0.0.0.0:3000:8080"
		return parts[1]
	default:
		return ""
	}
}

// CleanComposeFile removes the x-fnpack field and returns clean compose content
func CleanComposeFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read compose file: %w", err)
	}

	return CleanComposeContent(data)
}

// CleanComposeContent removes x-fnpack from compose content
func CleanComposeContent(data []byte) ([]byte, error) {
	// Parse as generic map to preserve all fields
	var content map[string]interface{}
	if err := yaml.Unmarshal(data, &content); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	// Remove x-fnpack field
	delete(content, "x-fnpack")

	// Marshal back to YAML
	cleanData, err := yaml.Marshal(content)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal clean yaml: %w", err)
	}

	return cleanData, nil
}

// GetManifestValue gets a value from the manifest with a default fallback
func GetManifestValue(manifest map[string]interface{}, key string, defaultValue string) string {
	if manifest == nil {
		return defaultValue
	}

	if value, ok := manifest[key]; ok {
		switch v := value.(type) {
		case string:
			return v
		case int, int64, float64:
			return fmt.Sprintf("%v", v)
		case bool:
			if v {
				return "yes"
			}
			return "no"
		default:
			return fmt.Sprintf("%v", v)
		}
	}

	return defaultValue
}
