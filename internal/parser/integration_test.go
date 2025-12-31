//go:build integration

package parser

import (
"fmt"
"testing"
)

func TestParseChromiumExample(t *testing.T) {
	compose, err := ParseComposeFile("../../../examples/Chromium/compose.yaml")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	fmt.Println("=== Manifest ===")
	for k, v := range compose.XFnpack.Manifest {
		fmt.Printf("%s = %v\n", k, v)
	}

	fmt.Println("\n=== Variables ===")
	vars := ExtractVariables(compose)
	fmt.Printf("ServiceName: %s\n", vars.ServiceName)
	fmt.Printf("ContainerName: %s\n", vars.ContainerName)
	fmt.Printf("FirstPort: %s\n", vars.FirstPort)

	// Verify expected values
	if vars.ServiceName != "chromium" {
		t.Errorf("Expected ServiceName 'chromium', got %s", vars.ServiceName)
	}
	if vars.ContainerName != "chromium" {
		t.Errorf("Expected ContainerName 'chromium', got %s", vars.ContainerName)
	}
	if vars.FirstPort != "3000" {
		t.Errorf("Expected FirstPort '3000', got %s", vars.FirstPort)
	}

	// Verify manifest values
	if compose.XFnpack.Manifest["appname"] != "docker-chromium" {
		t.Errorf("Expected appname 'docker-chromium', got %v", compose.XFnpack.Manifest["appname"])
	}
}
