package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// FnpackRunner handles execution of the fnpack CLI tool
type FnpackRunner struct {
	builder *Builder
}

// NewFnpackRunner creates a new FnpackRunner instance
func NewFnpackRunner(builder *Builder) *FnpackRunner {
	return &FnpackRunner{builder: builder}
}

// RunFnpack executes the fnpack build command to generate the .fpk file
// Returns the path to the generated .fpk file on success
func (r *FnpackRunner) RunFnpack() (string, error) {
	appDir := r.builder.GetAppDir()

	// Get absolute path for appDir
	absAppDir, err := filepath.Abs(appDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Find fnpack executable
	fnpackPath, err := r.findFnpack()
	if err != nil {
		return "", err
	}

	if r.builder.Verbose {
		fmt.Printf("Using fnpack: %s\n", fnpackPath)
		fmt.Printf("Building FPK from: %s\n", absAppDir)
	}

	// Get absolute path for output directory
	absOutputDir, err := filepath.Abs(r.builder.OutputDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute output path: %w", err)
	}

	// Execute fnpack build command
	// fnpack build <app_dir> - builds the fpk in the current directory
	cmd := exec.Command(fnpackPath, "build", absAppDir)
	cmd.Dir = absOutputDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("fnpack build failed: %w", err)
	}

	// Find the generated .fpk file
	fpkFile, err := r.findFpkFile()
	if err != nil {
		return "", err
	}

	if r.builder.Verbose {
		fmt.Printf("Generated FPK: %s\n", fpkFile)
	}

	return fpkFile, nil
}

// findFnpack searches for the fnpack executable in multiple locations
func (r *FnpackRunner) findFnpack() (string, error) {
	// 1. Check common locations relative to working directory first (prefer local bin)
	commonPaths := []string{
		"../bin/fnpack",
		"bin/fnpack",
		"../../bin/fnpack",
	}

	for _, path := range commonPaths {
		if absPath, err := filepath.Abs(path); err == nil {
			if _, err := os.Stat(absPath); err == nil {
				return absPath, nil
			}
		}
	}

	// 2. Check PATH
	if fnpackPath, err := exec.LookPath("fnpack"); err == nil {
		return fnpackPath, nil
	}

	// 3. Check system locations
	systemPaths := []string{
		"/usr/local/bin/fnpack",
	}

	for _, path := range systemPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("fnpack not found in PATH or common locations (bin/fnpack)")
}

// findFpkFile searches for the generated .fpk file in the output directory
// It looks for a file matching the app name pattern first, then falls back to any .fpk file
func (r *FnpackRunner) findFpkFile() (string, error) {
	appName := r.builder.AppName

	// First, try to find a .fpk file matching the app name
	// fnpack typically generates files like: appname-version.fpk or appname.fpk
	appPattern := filepath.Join(r.builder.OutputDir, appName+"*.fpk")
	matches, err := filepath.Glob(appPattern)
	if err != nil {
		return "", fmt.Errorf("failed to search for fpk file: %w", err)
	}

	if len(matches) > 0 {
		// Return the first match for this app
		return matches[0], nil
	}

	// Fallback: search for any .fpk file (for backwards compatibility)
	pattern := filepath.Join(r.builder.OutputDir, "*.fpk")
	matches, err = filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("failed to search for fpk file: %w", err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no .fpk file found in %s", r.builder.OutputDir)
	}

	// Return the first match
	return matches[0], nil
}

// BuildWithFnpack performs the complete build process including fnpack execution
func (b *Builder) BuildWithFnpack() (string, error) {
	// First, run the standard build process
	if err := b.Build(); err != nil {
		return "", err
	}

	// Then run fnpack to generate the .fpk file
	runner := NewFnpackRunner(b)
	return runner.RunFnpack()
}
