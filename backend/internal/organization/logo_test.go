package organization

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

func TestSanitizeLogoAcceptsPNG(t *testing.T) {
	var input bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 20, G: 80, B: 160, A: 255})
	if err := png.Encode(&input, img); err != nil {
		t.Fatal(err)
	}

	data, contentType, extension, err := sanitizeLogo(bytes.NewReader(input.Bytes()), int64(input.Len()))
	if err != nil {
		t.Fatalf("sanitizeLogo() error = %v", err)
	}
	if len(data) == 0 || contentType != "image/png" || extension != "png" {
		t.Fatalf("sanitizeLogo() = %d bytes, %q, %q", len(data), contentType, extension)
	}
}

func TestSanitizeLogoRejectsInvalidInput(t *testing.T) {
	if _, _, _, err := sanitizeLogo(strings.NewReader("not an image"), 12); !errors.Is(err, ErrInvalidLogo) {
		t.Fatalf("sanitizeLogo() error = %v, want ErrInvalidLogo", err)
	}
	if _, _, _, err := sanitizeLogo(bytes.NewReader(nil), maxLogoUploadBytes+1); !errors.Is(err, ErrLogoTooLarge) {
		t.Fatalf("sanitizeLogo() error = %v, want ErrLogoTooLarge", err)
	}
}
