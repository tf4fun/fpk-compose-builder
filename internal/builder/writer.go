package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fpk-compose-builder/internal/generator"
	"fpk-compose-builder/internal/parser"
)

// Writer handles writing generated files to the FPK directory
type Writer struct {
	builder *Builder
}

// NewWriter creates a new Writer instance
func NewWriter(builder *Builder) *Writer {
	return &Writer{builder: builder}
}

// WriteManifest writes the manifest file in key=value format
func (w *Writer) WriteManifest() error {
	content := generator.GenerateManifest(
		w.builder.Compose.XFnpack.Manifest,
		w.builder.Variables,
	)

	manifestPath := filepath.Join(w.builder.GetAppDir(), "manifest")
	if err := os.WriteFile(manifestPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	if w.builder.Verbose {
		fmt.Printf("Written: %s\n", manifestPath)
	}

	return nil
}

// WriteConfigs writes the privilege and resource configuration files (defaults)
// Only writes if not provided in x-fnpack files
func (w *Writer) WriteConfigs() error {
	files := w.builder.Compose.XFnpack.Files

	// Write privilege config if not provided
	if !w.hasFile(files, "config/privilege") {
		privilegeContent, err := generator.GeneratePrivilege(w.builder.Variables)
		if err != nil {
			return fmt.Errorf("failed to generate privilege config: %w", err)
		}

		privilegePath := filepath.Join(w.builder.GetAppDir(), "config", "privilege")
		if err := os.WriteFile(privilegePath, []byte(privilegeContent), 0644); err != nil {
			return fmt.Errorf("failed to write privilege config: %w", err)
		}

		if w.builder.Verbose {
			fmt.Printf("Written (default): %s\n", privilegePath)
		}
	}

	// Write resource config if not provided
	if !w.hasFile(files, "config/resource") {
		resourceContent, err := generator.GenerateResource(w.builder.Variables)
		if err != nil {
			return fmt.Errorf("failed to generate resource config: %w", err)
		}

		resourcePath := filepath.Join(w.builder.GetAppDir(), "config", "resource")
		if err := os.WriteFile(resourcePath, []byte(resourceContent), 0644); err != nil {
			return fmt.Errorf("failed to write resource config: %w", err)
		}

		if w.builder.Verbose {
			fmt.Printf("Written (default): %s\n", resourcePath)
		}
	}

	return nil
}

// WriteScript writes the cmd/main bash script and lifecycle scripts (defaults)
// Only writes if not provided in x-fnpack files
func (w *Writer) WriteScript() error {
	files := w.builder.Compose.XFnpack.Files

	// Write main script if not provided
	if !w.hasFile(files, "cmd/main") {
		content := generator.GenerateMainScript(w.builder.Variables)

		scriptPath := filepath.Join(w.builder.GetAppDir(), "cmd", "main")
		if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
			return fmt.Errorf("failed to write main script: %w", err)
		}

		if w.builder.Verbose {
			fmt.Printf("Written (default): %s\n", scriptPath)
		}
	}

	// Write lifecycle scripts if not provided
	lifecycleScripts := generator.GenerateLifecycleScripts()
	for name, content := range lifecycleScripts {
		filePath := "cmd/" + name
		if !w.hasFile(files, filePath) {
			scriptPath := filepath.Join(w.builder.GetAppDir(), "cmd", name)
			if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
				return fmt.Errorf("failed to write %s script: %w", name, err)
			}

			if w.builder.Verbose {
				fmt.Printf("Written (default): %s\n", scriptPath)
			}
		}
	}

	return nil
}

// WriteUIConfig writes the app/ui/config file (default)
// Only writes if not provided in x-fnpack files
func (w *Writer) WriteUIConfig() error {
	files := w.builder.Compose.XFnpack.Files

	if !w.hasFile(files, "app/ui/config") {
		content, err := generator.GenerateDefaultUIConfig(w.builder.Variables)
		if err != nil {
			return fmt.Errorf("failed to generate UI config: %w", err)
		}

		configPath := filepath.Join(w.builder.GetAppDir(), "app", "ui", "config")
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write UI config: %w", err)
		}

		if w.builder.Verbose {
			fmt.Printf("Written (default): %s\n", configPath)
		}
	}

	return nil
}

// CopyCompose copies the compose.yaml to app/docker/ with x-fnpack removed
func (w *Writer) CopyCompose() error {
	// Find the compose file
	composePath := filepath.Join(w.builder.InputDir, "compose.yaml")
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		composePath = filepath.Join(w.builder.InputDir, "docker-compose.yaml")
	}

	// Clean the compose content (remove x-fnpack)
	cleanContent, err := parser.CleanComposeFile(composePath)
	if err != nil {
		return fmt.Errorf("failed to clean compose file: %w", err)
	}

	// Write to app/docker/docker-compose.yaml
	destPath := filepath.Join(w.builder.GetAppDir(), "app", "docker", "docker-compose.yaml")
	if err := os.WriteFile(destPath, cleanContent, 0644); err != nil {
		return fmt.Errorf("failed to write compose file: %w", err)
	}

	if w.builder.Verbose {
		fmt.Printf("Written: %s\n", destPath)
	}

	return nil
}

// WriteLicense writes an empty LICENSE file
func (w *Writer) WriteLicense() error {
	licensePath := filepath.Join(w.builder.GetAppDir(), "LICENSE")
	if err := os.WriteFile(licensePath, []byte(""), 0644); err != nil {
		return fmt.Errorf("failed to write LICENSE: %w", err)
	}

	if w.builder.Verbose {
		fmt.Printf("Written: %s\n", licensePath)
	}

	return nil
}

// WriteCustomFiles writes all files defined in x-fnpack (except manifest)
// Files are written directly with variable replacement
func (w *Writer) WriteCustomFiles() error {
	files := w.builder.Compose.XFnpack.Files
	if files == nil {
		return nil
	}

	for filePath, content := range files {
		// Replace variables in content
		content = generator.ReplaceVariables(content, w.builder.Variables)

		// Create full path
		fullPath := filepath.Join(w.builder.GetAppDir(), filePath)

		// Ensure parent directory exists
		parentDir := filepath.Dir(fullPath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", filePath, err)
		}

		// Determine file permissions (executable for cmd/* files)
		perm := os.FileMode(0644)
		if strings.HasPrefix(filePath, "cmd/") {
			perm = 0755
		}

		// Write file
		if err := os.WriteFile(fullPath, []byte(content), perm); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filePath, err)
		}

		if w.builder.Verbose {
			fmt.Printf("Written: %s\n", fullPath)
		}
	}

	return nil
}

// hasFile checks if a file path exists in the files map
func (w *Writer) hasFile(files map[string]string, path string) bool {
	if files == nil {
		return false
	}
	_, exists := files[path]
	return exists
}
