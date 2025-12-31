package builder

import (
	"fmt"
	"os"
	"path/filepath"

	"fpk-compose-builder/internal/generator"
	"fpk-compose-builder/internal/parser"
)

// Builder handles the construction of FPK directory structure
type Builder struct {
	// InputDir is the directory containing compose.yaml and icon
	InputDir string

	// OutputDir is the directory where the FPK structure will be created
	OutputDir string

	// AppName is the application name (from manifest or service name)
	AppName string

	// Compose is the parsed compose file
	Compose *parser.ComposeFile

	// Variables contains extracted template variables
	Variables parser.Variables

	// Verbose enables detailed logging
	Verbose bool
}

// NewBuilder creates a new Builder instance
func NewBuilder(inputDir, outputDir string, verbose bool) *Builder {
	return &Builder{
		InputDir:  inputDir,
		OutputDir: outputDir,
		Verbose:   verbose,
	}
}

// Build orchestrates the complete FPK build process
func (b *Builder) Build() error {
	// Step 1: Parse compose file
	if err := b.parseCompose(); err != nil {
		return fmt.Errorf("failed to parse compose: %w", err)
	}

	// Step 2: Create directory structure
	if err := b.CreateDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Step 3: Write all files
	if err := b.writeAllFiles(); err != nil {
		return fmt.Errorf("failed to write files: %w", err)
	}

	// Step 4: Process icons
	if err := b.processIcons(); err != nil {
		return fmt.Errorf("failed to process icons: %w", err)
	}

	return nil
}


// parseCompose parses the compose file and extracts variables
func (b *Builder) parseCompose() error {
	composePath := filepath.Join(b.InputDir, "compose.yaml")

	// Try compose.yaml first, then docker-compose.yaml
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		composePath = filepath.Join(b.InputDir, "docker-compose.yaml")
		if _, err := os.Stat(composePath); os.IsNotExist(err) {
			return fmt.Errorf("compose.yaml or docker-compose.yaml not found in %s", b.InputDir)
		}
	}

	compose, err := parser.ParseComposeFile(composePath)
	if err != nil {
		return err
	}

	b.Compose = compose
	b.Variables = parser.ExtractVariables(compose)

	// Determine app name from manifest or service name
	b.AppName = generator.GetManifestAppname(compose.XFnpack.Manifest, b.Variables)

	if b.Verbose {
		fmt.Printf("Parsed compose file: %s\n", composePath)
		fmt.Printf("App name: %s\n", b.AppName)
		fmt.Printf("Service name: %s\n", b.Variables.ServiceName)
		fmt.Printf("Container name: %s\n", b.Variables.ContainerName)
		fmt.Printf("First port: %s\n", b.Variables.FirstPort)
	}

	return nil
}

// CreateDirectories creates the FPK directory structure
// Structure: app/docker, app/ui/images, cmd, config, wizard
func (b *Builder) CreateDirectories() error {
	appDir := filepath.Join(b.OutputDir, b.AppName)

	dirs := []string{
		filepath.Join(appDir, "app", "docker"),
		filepath.Join(appDir, "app", "ui", "images"),
		filepath.Join(appDir, "cmd"),
		filepath.Join(appDir, "config"),
		filepath.Join(appDir, "wizard"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		if b.Verbose {
			fmt.Printf("Created directory: %s\n", dir)
		}
	}

	return nil
}

// GetAppDir returns the full path to the app directory
func (b *Builder) GetAppDir() string {
	return filepath.Join(b.OutputDir, b.AppName)
}

// writeAllFiles writes all generated files to the FPK directory
func (b *Builder) writeAllFiles() error {
	writer := NewWriter(b)

	// Write manifest (always from YAML object -> key=value)
	if err := writer.WriteManifest(); err != nil {
		return err
	}

	// Write custom files from x-fnpack first (multi-line text -> file)
	// This allows custom files to override defaults
	if err := writer.WriteCustomFiles(); err != nil {
		return err
	}

	// Write default config files (privilege, resource) if not provided
	if err := writer.WriteConfigs(); err != nil {
		return err
	}

	// Write default cmd/main script if not provided
	if err := writer.WriteScript(); err != nil {
		return err
	}

	// Write default UI config if not provided
	if err := writer.WriteUIConfig(); err != nil {
		return err
	}

	// Copy compose file (cleaned, x-fnpack removed)
	if err := writer.CopyCompose(); err != nil {
		return err
	}

	// Write LICENSE file
	if err := writer.WriteLicense(); err != nil {
		return err
	}

	return nil
}

// processIcons finds and processes icon files
func (b *Builder) processIcons() error {
	iconHandler := NewIconHandler(b)
	return iconHandler.ProcessIcons()
}
