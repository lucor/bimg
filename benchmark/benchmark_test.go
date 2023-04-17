package benchmark_test

import (
	"os"
	"testing"

	"github.com/h2non/bimg"
)

var images = []struct {
	source string
}{
	{"3x3.jpg"},
	{"3x3greys.jpg"},
	{"northern_cardinal_bird.jpg"},
	{"parameter_trim.png"},
	{"test.avif"},
	{"test.gif"},
	{"test.png"},
	{"test.webp"},
	{"test2.heic"},
	{"test3.heic"},
	{"transparent.png"},
}

func Benchmark_RGBAPixels(b *testing.B) {
	for _, tt := range images {
		imagefile, err := os.ReadFile("../testdata/" + tt.source)
		if err != nil {
			b.Fatal(err)
		}
		b.Run(tt.source, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				bimg.RGBAPixels(imagefile)
			}
		})
	}
}
