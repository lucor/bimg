package bimg

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestVipsRead(t *testing.T) {
	files := []struct {
		name     string
		expected ImageType
	}{
		{"test.jpg", JPEG},
		{"test.png", PNG},
		{"test.webp", WEBP},
	}

	for _, file := range files {
		image, imageType, _ := vipsRead(readImage(file.name))
		if image == nil {
			t.Fatal("Empty image")
		}
		if imageType != file.expected {
			t.Fatal("Invalid image type")
		}
	}
}

func TestVipsSave(t *testing.T) {
	types := [...]ImageType{JPEG, PNG, WEBP}

	for _, typ := range types {
		image, _, _ := vipsRead(readImage("test.jpg"))
		options := vipsSaveOptions{Quality: 95, Type: typ, StripMetadata: true}

		buf, err := vipsSave(image, options)
		if err != nil {
			t.Fatalf("Cannot save the image as '%v'", ImageTypes[typ])
		}
		if len(buf) == 0 {
			t.Fatalf("Empty saved '%v' image", ImageTypes[typ])
		}
	}
}

func TestVipsSaveTiff(t *testing.T) {
	if !IsTypeSupportedSave(TIFF) {
		t.Skipf("Format %#v is not supported", ImageTypes[TIFF])
	}
	image, _, _ := vipsRead(readImage("test.jpg"))
	options := vipsSaveOptions{Quality: 95, Type: TIFF}
	buf, _ := vipsSave(image, options)

	if len(buf) == 0 {
		t.Fatalf("Empty saved '%v' image", ImageTypes[TIFF])
	}
}

func TestVipsSaveAvif(t *testing.T) {
	if !IsTypeSupportedSave(AVIF) {
		t.Skipf("Format %#v is not supported", ImageTypes[AVIF])
	}
	image, _, _ := vipsRead(readImage("test.jpg"))
	options := vipsSaveOptions{Quality: 95, Type: AVIF, Speed: 8}
	buf, err := vipsSave(image, options)
	if err != nil {
		t.Fatalf("Error saving image type %v: %v", ImageTypes[AVIF], err)
	}

	if len(buf) == 0 {
		t.Fatalf("Empty saved '%v' image", ImageTypes[AVIF])
	}
}

func TestVipsRotate(t *testing.T) {
	files := []struct {
		name   string
		rotate Angle
	}{
		{"test.jpg", D90},
		{"test_square.jpg", D45},
	}

	for _, file := range files {
		image, _, _ := vipsRead(readImage(file.name))

		newImg, err := vipsRotate(image, file.rotate, nil)
		if err != nil {
			t.Fatal("Cannot rotate the image")
		}

		buf, _ := vipsSave(newImg, vipsSaveOptions{Quality: 95})
		if len(buf) == 0 {
			t.Fatal("Empty image")
		}
	}
}

func TestVipsAutoRotate(t *testing.T) {
	if VipsMajorVersion <= 8 && VipsMinorVersion < 10 {
		t.Skip("Skip test in libvips < 8.10")
		return
	}

	files := []struct {
		name        string
		orientation int
	}{
		{"test.jpg", 0},
		{"test_exif.jpg", 0},
		{"exif/Landscape_1.jpg", 0},
		{"exif/Landscape_2.jpg", 0},
		{"exif/Landscape_3.jpg", 0},
		{"exif/Landscape_4.jpg", 0},
		{"exif/Landscape_5.jpg", 5},
		{"exif/Landscape_6.jpg", 0},
		{"exif/Landscape_7.jpg", 7},
	}

	for _, file := range files {
		image, _, _ := vipsRead(readImage(file.name))

		newImg, err := vipsAutoRotate(image)
		if err != nil {
			t.Fatal("Cannot auto rotate the image")
		}

		orientation := vipsExifOrientation(newImg)
		if orientation != file.orientation {
			t.Fatalf("Invalid image orientation: %d != %d", orientation, file.orientation)
		}

		buf, _ := vipsSave(newImg, vipsSaveOptions{Quality: 95})
		if len(buf) == 0 {
			t.Fatal("Empty image")
		}
	}
}

func TestVipsZoom(t *testing.T) {
	image, _, _ := vipsRead(readImage("test.jpg"))

	newImg, err := vipsZoom(image, 1)
	if err != nil {
		t.Fatal("Cannot save the image")
	}

	buf, _ := vipsSave(newImg, vipsSaveOptions{Quality: 95})
	if len(buf) == 0 {
		t.Fatal("Empty image")
	}
}

func TestVipsWatermark(t *testing.T) {
	image, _, _ := vipsRead(readImage("test.jpg"))

	watermark := Watermark{
		Text:       "Copy me if you can",
		Font:       "sans bold 12",
		Opacity:    0.5,
		Width:      200,
		DPI:        100,
		Margin:     100,
		Background: Color{255, 255, 255},
	}

	newImg, err := vipsWatermark(image, watermark)
	if err != nil {
		t.Errorf("Cannot add watermark: %s", err)
	}

	buf, _ := vipsSave(newImg, vipsSaveOptions{Quality: 95})
	if len(buf) == 0 {
		t.Fatal("Empty image")
	}
}

func TestVipsWatermarkWithImage(t *testing.T) {
	image, _, _ := vipsRead(readImage("test.jpg"))

	watermark := readImage("transparent.png")

	options := WatermarkImage{Left: 100, Top: 100, Opacity: 1.0, Buf: watermark}
	newImg, err := vipsDrawWatermark(image, options)
	if err != nil {
		t.Errorf("Cannot add watermark: %s", err)
	}

	buf, _ := vipsSave(newImg, vipsSaveOptions{Quality: 95})
	if len(buf) == 0 {
		t.Fatal("Empty image")
	}
}

func TestVipsImageType(t *testing.T) {
	imgType := vipsImageType(readImage("test.jpg"))
	if imgType != JPEG {
		t.Fatal("Invalid image type")
	}
}

func TestVipsImageTypeInvalid(t *testing.T) {
	imgType := vipsImageType([]byte("vip"))
	if imgType != UNKNOWN {
		t.Fatal("Invalid image type")
	}
}

func TestVipsMemory(t *testing.T) {
	mem := VipsMemory()

	if mem.Memory < 1024 {
		t.Fatal("Invalid memory")
	}
	if mem.Allocations == 0 {
		t.Fatal("Invalid memory allocations")
	}
}

func TestVipsExifShort(t *testing.T) {
	tt := []struct {
		input    string
		expected string
	}{
		{
			input:    `( ()`,
			expected: `(`,
		},
		{
			input:    ` ()`,
			expected: ` ()`,
		},
		{
			input:    `sRGB`,
			expected: `sRGB`,
		},
	}

	for _, tc := range tt {
		got := vipsExifShort(tc.input)
		if got != tc.expected {
			t.Fatalf("expected: %s; got: %s", tc.expected, got)
		}
	}
}

func readImage(file string) []byte {
	img, _ := os.Open(path.Join("testdata", file))
	buf, _ := ioutil.ReadAll(img)
	defer img.Close()
	return buf
}

// printfImageAsRGBA : print images bytes in hex
func printfImageAsRGBA(t *testing.T, img []uint8, w int) {
	for i := 0; i < len(img); i += 4 {
		if ((i)%(w*4)) == 0 && i != 0 {
			fmt.Printf("\n")
		}
		t.Logf("%2x %2x %2x (%2x) | ", img[i], img[i+1], img[i+2], img[i+3])
	}
	t.Log("\n")
}

func Test_RGBAPixels(t *testing.T) {

	// ./testdata/northern_cardinal_bird.jpg JPEG 1920x1080 1920x1080+0+0 8-bit sRGB 828936B
	imagefile, err := os.ReadFile("testdata/northern_cardinal_bird.jpg")
	if err != nil {
		t.Fatal(err)
	}

	pix, width, height, err := RGBAPixels(imagefile)

	if err != nil {
		t.Fatalf("could not get RGBAPixels for origin: %v", err)
	}

	if width != 1920 {
		t.Fatalf("wrong width %d\n", width)
	}

	if height != 1080 {
		t.Fatalf("wrong height %d\n", height)
	}

	if height*width*4 != len(pix) {
		t.Fatalf("wrong slice len %d\n", len(pix))
	}

	t.Logf("TestRGBAPixels returned %d len rgba byte slice, width %d, height %d\n", len(pix), width, height)
}

// Test_RGBAPixelsFormatBands
func Test_RGBAPixelsFormatBands(t *testing.T) {

	tests := []struct {
		name       string
		source     string
		wantPixels []uint8
	}{
		{
			name:   "band 1",
			source: "3x3greys.jpg",
			wantPixels: []uint8{
				255, 255, 255, 255, //ffffff
				210, 210, 210, 255, //d2d2d2
				196, 196, 196, 255, //c4c4c4
				188, 188, 188, 255, //bcbcbc
				149, 149, 149, 255, //959595
				81, 81, 81, 255, //515151
				0, 0, 0, 255, //000000
				7, 7, 7, 255, //070707
				9, 9, 9, 255, //090909
			},
		},
		{
			name:   "band 3",
			source: "3x3.jpg",
			wantPixels: []uint8{
				255, 255, 255, 255, //ffffff
				210, 210, 210, 255, //d2d2d2
				196, 196, 196, 255, //c4c4c4
				188, 188, 188, 255, //bcbcbc
				149, 149, 149, 255, //959595
				81, 81, 81, 255, //515151
				0, 0, 0, 255, //000000
				7, 7, 7, 255, //070707
				9, 9, 9, 255, //090909
			},
		},
		{
			name:   "band 4",
			source: "3x3.png",
			wantPixels: []uint8{
				49, 49, 189, 31,
				0, 0, 0, 0,
				0, 0, 0, 0,
				45, 45, 177, 54,
				44, 44, 174, 51,
				43, 43, 169, 54,
				42, 42, 167, 85,
				39, 39, 154, 82,
				40, 40, 158, 82,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imagefile, err := os.ReadFile("testdata/" + tt.source)
			if err != nil {
				t.Fatal(err)
			}

			pix, width, height, err := RGBAPixels(imagefile)

			if err != nil {
				t.Fatalf("could not get RGBAPixels for origin: %v", err)
			}

			if width != 3 {
				t.Fatalf("wrong width %d\n", width)
			}

			if height != 3 {
				t.Fatalf("wrong height %d\n", height)
			}

			if height*width*4 != len(pix) {
				t.Fatalf("wrong slice len %d\n", len(pix))
			}

			t.Logf("TestRGBAPixels returned %d len rgba byte slice, width %d, height %d\n", len(pix), width, height)

			if !bytes.Equal(tt.wantPixels, pix) {
				printfImageAsRGBA(t, pix, width)
				t.Fatal("invalid pixel data")
				t.Logf("Image bytes: \n")
			}
		})
	}
}
