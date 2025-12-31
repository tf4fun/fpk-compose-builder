package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"fpk-compose-builder/internal/builder"
	"fpk-compose-builder/internal/generator"
)

var (
	// Version information (can be set at build time)
	version = "dev"

	// CLI flags
	inputDir  string
	outputDir string
	verbose   bool
	skipFnpack bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "fpk-compose-builder",
	Short: "Build fnOS FPK packages from docker-compose files",
	Long: `fpk-compose-builder converts docker-compose.yaml files with x-fnpack 
extension fields into fnOS FPK application packages.

It parses the compose file, extracts metadata, generates required 
configuration files, and optionally invokes fnpack to build the final .fpk file.`,
	Version: version,
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build FPK package from compose file",
	Long: `Build an FPK package from a directory containing compose.yaml and optional icon.png.

The compose.yaml file should contain an x-fnpack section with manifest metadata
and optional wizard/UI configurations. The tool will:

1. Parse the compose file and extract x-fnpack configuration
2. Generate manifest, config files, and scripts
3. Process icons (resize to required dimensions)
4. Optionally invoke fnpack to create the final .fpk file

Example:
  fpk-compose-builder build -i examples/Chromium -o dist/`,
	RunE: runBuild,
}

func init() {
	// Add build command to root
	rootCmd.AddCommand(buildCmd)

	// Build command flags
	buildCmd.Flags().StringVarP(&inputDir, "input", "i", ".", "Input directory containing compose.yaml and icon.png")
	buildCmd.Flags().StringVarP(&outputDir, "output", "o", "./dist", "Output directory for generated FPK structure")
	buildCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	buildCmd.Flags().BoolVar(&skipFnpack, "skip-fnpack", false, "Skip fnpack build step (only generate directory structure)")
}


func runBuild(cmd *cobra.Command, args []string) error {
	// Validate input directory exists
	if _, err := os.Stat(inputDir); os.IsNotExist(err) {
		return fmt.Errorf("input directory does not exist: %s", inputDir)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if verbose {
		fmt.Printf("Input directory: %s\n", inputDir)
		fmt.Printf("Output directory: %s\n", outputDir)
		fmt.Println("Starting build process...")
	}

	// Create builder and run the build process
	b := builder.NewBuilder(inputDir, outputDir, verbose)

	if skipFnpack {
		// Only generate directory structure, skip fnpack
		if err := b.Build(); err != nil {
			return fmt.Errorf("build failed: %w", err)
		}

		fmt.Printf("✓ FPK directory structure generated at: %s/%s\n", outputDir, b.AppName)
		printBuildSummary(b)
	} else {
		// Full build with fnpack
		fpkFile, err := b.BuildWithFnpack()
		if err != nil {
			return fmt.Errorf("build failed: %w", err)
		}

		fmt.Printf("✓ FPK package built successfully: %s\n", fpkFile)
		printBuildSummary(b)
	}

	return nil
}

func printBuildSummary(b *builder.Builder) {
	if b.Compose == nil {
		return
	}

	appName := b.AppName
	version := generator.GetManifestVersion(b.Compose.XFnpack.Manifest)

	fmt.Println("\nBuild Summary:")
	fmt.Printf("  App Name:    %s\n", appName)
	fmt.Printf("  Version:     %s\n", version)
	fmt.Printf("  Service:     %s\n", b.Variables.ServiceName)
	if b.Variables.FirstPort != "" {
		fmt.Printf("  Port:        %s\n", b.Variables.FirstPort)
	}
}
