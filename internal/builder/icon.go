package builder

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

// IconHandler handles icon processing for FPK packages
type IconHandler struct {
	builder *Builder
}

// NewIconHandler creates a new IconHandler instance
func NewIconHandler(builder *Builder) *IconHandler {
	return &IconHandler{builder: builder}
}

// ProcessIcons finds, resizes, and copies icons to the FPK directory
func (h *IconHandler) ProcessIcons() error {
	// Find icon in input directory
	iconPath, err := h.FindIcon()
	if err != nil {
		if h.builder.Verbose {
			fmt.Printf("No icon found: %v, using default\n", err)
		}
		// No icon found, skip icon processing
		return nil
	}

	if h.builder.Verbose {
		fmt.Printf("Found icon: %s\n", iconPath)
	}

	// Load the source image
	srcImg, err := imaging.Open(iconPath)
	if err != nil {
		return fmt.Errorf("failed to open icon: %w", err)
	}

	// Pad to square if not already square
	srcImg = h.squareImage(srcImg)

	if h.builder.Verbose {
		bounds := srcImg.Bounds()
		fmt.Printf("Icon prepared: %dx%d (squared)\n", bounds.Dx(), bounds.Dy())
	}

	// Generate and copy icons
	if err := h.CopyIcons(srcImg); err != nil {
		return err
	}

	return nil
}

// FindIcon searches for the first .png file in the input directory
func (h *IconHandler) FindIcon() (string, error) {
	entries, err := os.ReadDir(h.builder.InputDir)
	if err != nil {
		return "", fmt.Errorf("failed to read input directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".png") {
			return filepath.Join(h.builder.InputDir, name), nil
		}
	}

	return "", fmt.Errorf("no .png file found in %s", h.builder.InputDir)
}

// squareImage pads a non-square image to make it square, centering the original image
// The background is transparent
func (h *IconHandler) squareImage(src image.Image) image.Image {
	bounds := src.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	// If already square, return as-is
	if srcWidth == srcHeight {
		return src
	}

	// Determine the size of the square (use the larger dimension)
	size := srcWidth
	if srcHeight > srcWidth {
		size = srcHeight
	}

	// Create a new transparent square image
	dst := image.NewRGBA(image.Rect(0, 0, size, size))

	// Fill with transparent background
	draw.Draw(dst, dst.Bounds(), &image.Uniform{color.Transparent}, image.Point{}, draw.Src)

	// Calculate offset to center the original image
	offsetX := (size - srcWidth) / 2
	offsetY := (size - srcHeight) / 2

	// Draw the source image centered on the square canvas
	dstRect := image.Rect(offsetX, offsetY, offsetX+srcWidth, offsetY+srcHeight)
	draw.Draw(dst, dstRect, src, bounds.Min, draw.Over)

	return dst
}

// ResizeIcon resizes an image to the specified dimensions
// Uses Lanczos resampling for high-quality results
func (h *IconHandler) ResizeIcon(src image.Image, width, height int) image.Image {
	return imaging.Resize(src, width, height, imaging.Lanczos)
}

// CopyIcons generates and copies icons to all required locations
// Generates: ICON.PNG (64x64), ICON_256.PNG (256x256)
// Also copies to: app/ui/images/icon-64.png, app/ui/images/icon-256.png
func (h *IconHandler) CopyIcons(srcImg image.Image) error {
	appDir := h.builder.GetAppDir()

	// Define icon sizes and destinations
	// Note: app/ui/images uses icon-{size}.png format (hyphen, lowercase extension)
	icons := []struct {
		width    int
		height   int
		destPath string
	}{
		{64, 64, filepath.Join(appDir, "ICON.PNG")},
		{256, 256, filepath.Join(appDir, "ICON_256.PNG")},
		{64, 64, filepath.Join(appDir, "app", "ui", "images", "icon-64.png")},
		{256, 256, filepath.Join(appDir, "app", "ui", "images", "icon-256.png")},
	}

	for _, icon := range icons {
		// Resize the image
		resized := h.ResizeIcon(srcImg, icon.width, icon.height)

		// Save the resized image
		if err := h.saveIcon(resized, icon.destPath); err != nil {
			return fmt.Errorf("failed to save icon %s: %w", icon.destPath, err)
		}

		if h.builder.Verbose {
			fmt.Printf("Written: %s (%dx%d)\n", icon.destPath, icon.width, icon.height)
		}
	}

	return nil
}

// saveIcon saves an image to the specified path as PNG
func (h *IconHandler) saveIcon(img image.Image, destPath string) error {
	// Ensure parent directory exists
	parentDir := filepath.Dir(destPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the output file
	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	// Encode and save as PNG
	if err := png.Encode(outFile, img); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}
