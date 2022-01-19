package zplgfa

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"strings"
	"testing"
)

//go:embed label.png
var imgPNG []byte

//go:embed label.jpg
var imgJPG []byte

//go:embed label.expected.txt
var labelExpected string

func init() {
	// windows specific, reading from file, line endings are different
	labelExpected = strings.ReplaceAll(labelExpected, "\r\n", "\n")
}

func TestPNGFullProcessing(t *testing.T) {
	buf := bytes.NewBuffer(imgPNG)
	img, _, err := image.Decode(buf)
	if err != nil {
		t.Fatal(err)
	}
	gfimg := ConvertToZPL(img, CompressedASCII)
	if gfimg != labelExpected {
		t.Fatalf("unexpected output, got: %s", gfimg)
	}
}

// func BenchmarkColorConversion(b *testing.B) {
// 	for n := 0; n < b.N; n++ {
// 		white := color.White
// 		otherwhite := color.NRGBA64Model.Convert(white).(color.NRGBA64)
// 		_ = otherwhite
// 	}
// }

// func BenchmarkColorConversionCustom(b *testing.B) {
// 	for n := 0; n < b.N; n++ {
// 		white := color.White
// 		otherwhite := toNRGBA(white)
// 		_ = otherwhite
// 	}
// }

func BenchmarkPNGFullProcessing(b *testing.B) {
	for n := 0; n < b.N; n++ {
		buf := bytes.NewBuffer(imgPNG)
		img, _, err := image.Decode(buf)
		if err != nil {
			b.Fatal(err)
		}
		gfimg := ConvertToZPL(img, CompressedASCII)
		if len(gfimg) < 1 {
			b.Fatalf("expected an img")
		}
	}
}

// func BenchmarkJPGFullProcessing(b *testing.B) {
// 	for n := 0; n < b.N; n++ {
// 		buf := bytes.NewBuffer(imgJPG)
// 		img, _, err := image.Decode(buf)
// 		if err != nil {
// 			b.Fatal(err)
// 		}
// 		flat := FlattenImage(img)
// 		gfimg := ConvertToZPL(flat, CompressedASCII)
// 		if len(gfimg) < 1 {
// 			b.Fatalf("expected an img")
// 		}
// 	}
// }

// func BenchmarkPNGFlatten(b *testing.B) {
// 	for n := 0; n < b.N; n++ {
// 		buf := bytes.NewBuffer(imgPNG)
// 		img, _, err := image.Decode(buf)
// 		if err != nil {
// 			b.Fatal(err)
// 		}
// 		flat := FlattenImage(img)
// 		size := flat.Rect.Size()
// 		if size.X*size.Y < 1 {
// 			b.Fatalf("expected an img")
// 		}
// 	}
// }

// func BenchmarkJPGFlatten(b *testing.B) {
// 	for n := 0; n < b.N; n++ {
// 		buf := bytes.NewBuffer(imgJPG)
// 		img, _, err := image.Decode(buf)
// 		if err != nil {
// 			b.Fatal(err)
// 		}
// 		flat := FlattenImage(img)
// 		size := flat.Rect.Size()
// 		if size.X*size.Y < 1 {
// 			b.Fatalf("expected an img")
// 		}
// 	}
// }

func TestFlatten(t *testing.T) {
	tests := []struct {
		modelName string
		input     color.Color
		expected  color.Gray16
	}{
		{modelName: "Gray16", input: color.White, expected: color.White},
		{modelName: "Gray16", input: color.Black, expected: color.Black},
		{modelName: "RGBA", input: color.RGBA{0x98, 0x33, 0x87, 0x98}, expected: color.Gray16{0x7e95}},
		{modelName: "RGBA", input: color.RGBA{0x7F, 0x27, 0x33, 0xC2}, expected: color.Gray16{0x50ad}},
		{modelName: "RGBA", input: color.RGBA{0x11, 0x15, 0x50, 0x89}, expected: color.Gray16{0x7f7f}},
		{modelName: "RGBA", input: color.RGBA{0x94, 0xE0, 0x98, 0xE0}, expected: color.Gray16{0xc556}},
		{modelName: "RGBA", input: color.RGBA{0x1D, 0xCC, 0x37, 0xCC}, expected: color.Gray16{0x8134}},
		{modelName: "RGBA", input: color.RGBA{0x21, 0x27, 0x9F, 0x9F}, expected: color.Gray16{0x74dd}},
		{modelName: "RGBA", input: color.RGBA{0x88, 0xC8, 0x08, 0xF4}, expected: color.Gray16{0xa311}},
		{modelName: "RGBA", input: color.RGBA{0x25, 0xDA, 0x57, 0xDA}, expected: color.Gray16{0x902b}},
		{modelName: "RGBA", input: color.RGBA{0x64, 0xD0, 0x51, 0xD0}, expected: color.Gray16{0x99f4}},
		{modelName: "RGBA", input: color.RGBA{0xB2, 0xA0, 0x18, 0xC1}, expected: color.Gray16{0x8b45}},
		{modelName: "RGBA", input: color.RGBA{0x84, 0x6B, 0xF3, 0xF3}, expected: color.Gray16{0x81a7}},
		{modelName: "RGBA", input: color.RGBA{0x2C, 0x9B, 0xC5, 0xC5}, expected: color.Gray16{0x6e97}},
		{modelName: "RGBA", input: color.RGBA{0x90, 0x7F, 0x60, 0xE7}, expected: color.Gray16{0x8195}},
		{modelName: "RGBA", input: color.RGBA{0x5E, 0x93, 0xB6, 0xDA}, expected: color.Gray16{0x8397}},
		{modelName: "RGBA", input: color.RGBA{0xCD, 0x4C, 0x56, 0xCD}, expected: color.Gray16{0x69ad}},
		{modelName: "RGBA", input: color.RGBA{0x66, 0x2C, 0x4E, 0x66}, expected: color.Gray16{0xa3e4}},
		{modelName: "RGBA", input: color.RGBA{0x78, 0x46, 0x5F, 0xEE}, expected: color.Gray16{0x5bd6}},
		{modelName: "RGBA", input: color.RGBA{0xBB, 0xA7, 0x7E, 0xBB}, expected: color.Gray16{0x9250}},
		{modelName: "RGBA", input: color.RGBA{0xC6, 0xAE, 0xBA, 0xC6}, expected: color.Gray16{0xbddc}},
		{modelName: "RGBA", input: color.RGBA{0x85, 0xC6, 0x9E, 0xC9}, expected: color.Gray16{0xa49f}},
		{modelName: "NRGBA", input: color.NRGBA{0xE6, 0x2D, 0xB0, 0xF3}, expected: color.Gray16{0x708a}},
		{modelName: "NRGBA", input: color.NRGBA{0x54, 0xC8, 0xAC, 0xDB}, expected: color.Gray16{0x8b38}},
		{modelName: "NRGBA", input: color.NRGBA{0x17, 0xDD, 0x68, 0xB2}, expected: color.Gray16{0x667b}},
		{modelName: "NRGBA", input: color.NRGBA{0x1B, 0x28, 0xC6, 0xFD}, expected: color.Gray16{0x351a}},
		{modelName: "NRGBA", input: color.NRGBA{0x93, 0xAE, 0x3B, 0xC9}, expected: color.Gray16{0x768d}},
		{modelName: "NRGBA", input: color.NRGBA{0x23, 0x36, 0xB2, 0x9D}, expected: color.Gray16{0x722f}},
		{modelName: "NRGBA", input: color.NRGBA{0x45, 0x50, 0xDE, 0xEB}, expected: color.Gray16{0x59be}},
		{modelName: "NRGBA", input: color.NRGBA{0x14, 0xAE, 0xDC, 0x5A}, expected: color.Gray16{0xb2a6}},
		{modelName: "NRGBA", input: color.NRGBA{0x54, 0x01, 0x9D, 0x80}, expected: color.Gray16{0x7f7f}},
		{modelName: "NRGBA", input: color.NRGBA{0x8A, 0x1A, 0xBC, 0x9F}, expected: color.Gray16{0x6d9a}},
		{modelName: "NRGBA", input: color.NRGBA{0xCC, 0x69, 0x13, 0xCD}, expected: color.Gray16{0x8343}},
		{modelName: "NRGBA", input: color.NRGBA{0x44, 0x36, 0xA5, 0xD2}, expected: color.Gray16{0x36ee}},
		{modelName: "NRGBA", input: color.NRGBA{0xC4, 0x79, 0x10, 0x8F}, expected: color.Gray16{0x7873}},
		{modelName: "NRGBA", input: color.NRGBA{0xCD, 0xEF, 0x45, 0x0A}, expected: color.Gray16{0xf5f5}},
		{modelName: "NRGBA", input: color.NRGBA{0x62, 0xEF, 0x1F, 0xF7}, expected: color.Gray16{0xa83e}},
		{modelName: "NRGBA", input: color.NRGBA{0x55, 0x99, 0x2B, 0xAB}, expected: color.Gray16{0x5f77}},
		{modelName: "NRGBA", input: color.NRGBA{0x3E, 0x7C, 0x44, 0x67}, expected: color.Gray16{0x9be5}},
		{modelName: "NRGBA", input: color.NRGBA{0xE5, 0x48, 0xE5, 0x4E}, expected: color.Gray16{0xb6e3}},
		{modelName: "NRGBA", input: color.NRGBA{0xB0, 0x17, 0xC6, 0x07}, expected: color.Gray16{0xf8f8}},
		{modelName: "NRGBA", input: color.NRGBA{0xB9, 0xD1, 0x3F, 0xF7}, expected: color.Gray16{0xb333}},
		{modelName: "RGBA64", input: color.RGBA64{0x1FBB, 0x8D0A, 0x5DDE, 0x974D}, expected: color.Gray16{0x7ba3}},
		{modelName: "RGBA64", input: color.RGBA64{0x94C8, 0xEB90, 0xBE5, 0xEB90}, expected: color.Gray16{0xb3e8}},
		{modelName: "RGBA64", input: color.RGBA64{0xDAC1, 0xD25B, 0x6A6B, 0xDAC1}, expected: color.Gray16{0xb3b5}},
		{modelName: "RGBA64", input: color.RGBA64{0xADE7, 0xD7EE, 0x976A, 0xD7EE}, expected: color.Gray16{0xb655}},
		{modelName: "RGBA64", input: color.RGBA64{0xE368, 0x376D, 0xC121, 0xE368}, expected: color.Gray16{0x7c71}},
		{modelName: "RGBA64", input: color.RGBA64{0x988C, 0x9C5D, 0xA827, 0xA827}, expected: color.Gray16{0x7861}},
		{modelName: "RGBA64", input: color.RGBA64{0x8866, 0x985A, 0x4077, 0x985A}, expected: color.Gray16{0x7a59}},
		{modelName: "RGBA64", input: color.RGBA64{0x6D40, 0x765E, 0xB971, 0xB971}, expected: color.Gray16{0x61a4}},
		{modelName: "RGBA64", input: color.RGBA64{0x5DBA, 0xE518, 0x4E4, 0xE518}, expected: color.Gray16{0xa229}},
		{modelName: "RGBA64", input: color.RGBA64{0x6B66, 0xE2CC, 0xA0E1, 0xE2CC}, expected: color.Gray16{0xb0f2}},
		{modelName: "RGBA64", input: color.RGBA64{0xDFE0, 0xDCDE, 0x3797, 0xFC74}, expected: color.Gray16{0xca4a}},
		{modelName: "RGBA64", input: color.RGBA64{0xF7B2, 0x63E, 0x802E, 0xF7B2}, expected: color.Gray16{0x5e2e}},
		{modelName: "RGBA64", input: color.RGBA64{0x947A, 0x5594, 0xDC72, 0xDC72}, expected: color.Gray16{0x7b09}},
		{modelName: "RGBA64", input: color.RGBA64{0x25FD, 0x7996, 0xE4FF, 0xE4FF}, expected: color.Gray16{0x7612}},
		{modelName: "RGBA64", input: color.RGBA64{0x3D91, 0x845E, 0x3B4A, 0xB175}, expected: color.Gray16{0x65b4}},
		{modelName: "RGBA64", input: color.RGBA64{0xD29, 0xB9EE, 0x518D, 0xD70D}, expected: color.Gray16{0x880e}},
		{modelName: "RGBA64", input: color.RGBA64{0x92CC, 0xCA88, 0x4776, 0xCA88}, expected: color.Gray16{0x94c3}},
		{modelName: "RGBA64", input: color.RGBA64{0x601F, 0xCA2D, 0x440F, 0xCA2D}, expected: color.Gray16{0x9cbe}},
		{modelName: "RGBA64", input: color.RGBA64{0x5AFD, 0xF3D2, 0xA7F4, 0xF3D2}, expected: color.Gray16{0xb97d}},
		{modelName: "RGBA64", input: color.RGBA64{0x2A83, 0x5BEB, 0xD7EE, 0xD7EE}, expected: color.Gray16{0x62e3}},
		{modelName: "NRGBA64", input: color.NRGBA64{0x11B4, 0xF7E4, 0x6B01, 0x933E}, expected: color.Gray16{0x77ac}},
		{modelName: "NRGBA64", input: color.NRGBA64{0x8D85, 0x1688, 0x8437, 0xAAE3}, expected: color.Gray16{0x6ca3}},
		{modelName: "NRGBA64", input: color.NRGBA64{0xD759, 0x7DFC, 0x5184, 0x58FB}, expected: color.Gray16{0xb47c}},
		{modelName: "NRGBA64", input: color.NRGBA64{0xC86C, 0x4212, 0x8189, 0x5916}, expected: color.Gray16{0xb399}},
		{modelName: "NRGBA64", input: color.NRGBA64{0x1391, 0x87E, 0x120D, 0xE2B2}, expected: color.Gray16{0x1f1f}},
		{modelName: "NRGBA64", input: color.NRGBA64{0xA7B4, 0xF9B4, 0x4255, 0xEA4C}, expected: color.Gray16{0xb2f1}},
		{modelName: "NRGBA64", input: color.NRGBA64{0xBC0A, 0x709C, 0x23B5, 0x4564}, expected: color.Gray16{0xbc3a}},
		{modelName: "NRGBA64", input: color.NRGBA64{0x73B4, 0x3249, 0x5F28, 0x4573}, expected: color.Gray16{0xbbe3}},
		{modelName: "NRGBA64", input: color.NRGBA64{0xC65D, 0xEB36, 0x22DD, 0x9B2A}, expected: color.Gray16{0x7251}},
		{modelName: "NRGBA64", input: color.NRGBA64{0xB9E2, 0x2FD7, 0x7C24, 0xF0CB}, expected: color.Gray16{0x5cee}},
		{modelName: "NRGBA64", input: color.NRGBA64{0xA428, 0x7BA, 0xA772, 0x3D0B}, expected: color.Gray16{0xc67d}},
		{modelName: "NRGBA64", input: color.NRGBA64{0xF44A, 0x2717, 0x8222, 0xF36C}, expected: color.Gray16{0x6c0a}},
		{modelName: "NRGBA64", input: color.NRGBA64{0x2D11, 0xFFF7, 0x77FF, 0xA3F0}, expected: color.Gray16{0x7398}},
		{modelName: "NRGBA64", input: color.NRGBA64{0x6B69, 0x7656, 0x2539, 0x20BE}, expected: color.Gray16{0xdfdf}},
		{modelName: "NRGBA64", input: color.NRGBA64{0x5000, 0x1BBF, 0x6CAD, 0x228}, expected: color.Gray16{0xfdfd}},
		{modelName: "NRGBA64", input: color.NRGBA64{0x139B, 0x2BA6, 0x23E, 0x67CB}, expected: color.Gray16{0x9d9f}},
		{modelName: "NRGBA64", input: color.NRGBA64{0xD12D, 0x7903, 0x8DB5, 0x63FF}, expected: color.Gray16{0x9ece}},
		{modelName: "NRGBA64", input: color.NRGBA64{0xEC74, 0x4FD9, 0x68B2, 0x74FE}, expected: color.Gray16{0xa5ab}},
		{modelName: "NRGBA64", input: color.NRGBA64{0x31B5, 0x87, 0x71C2, 0xC86B}, expected: color.Gray16{0x40f1}},
		{modelName: "NRGBA64", input: color.NRGBA64{0x6B6, 0xF2DF, 0x5B2B, 0x9A01}, expected: color.Gray16{0x709a}},
		{modelName: "Gray16", input: color.Gray16{0xFE4E}, expected: color.Gray16{0xfefe}},
		{modelName: "Gray16", input: color.Gray16{0xC464}, expected: color.Gray16{0xc4c4}},
		{modelName: "Gray16", input: color.Gray16{0xDC88}, expected: color.Gray16{0xdcdc}},
		{modelName: "Gray16", input: color.Gray16{0xC83D}, expected: color.Gray16{0xc8c8}},
		{modelName: "Gray16", input: color.Gray16{0x4331}, expected: color.Gray16{0x4343}},
		{modelName: "Gray16", input: color.Gray16{0xCDF4}, expected: color.Gray16{0xcdcd}},
		{modelName: "Gray16", input: color.Gray16{0xE305}, expected: color.Gray16{0xe3e3}},
		{modelName: "Gray16", input: color.Gray16{0x5C9}, expected: color.Gray16{0x505}},
		{modelName: "Gray16", input: color.Gray16{0x28B4}, expected: color.Gray16{0x2828}},
		{modelName: "Gray16", input: color.Gray16{0x423F}, expected: color.Gray16{0x4242}},
		{modelName: "Gray16", input: color.Gray16{0x7B98}, expected: color.Gray16{0x7b7b}},
		{modelName: "Gray16", input: color.Gray16{0x3380}, expected: color.Gray16{0x3333}},
		{modelName: "Gray16", input: color.Gray16{0x75CB}, expected: color.Gray16{0x7575}},
		{modelName: "Gray16", input: color.Gray16{0x39AC}, expected: color.Gray16{0x3939}},
		{modelName: "Gray16", input: color.Gray16{0x2BA6}, expected: color.Gray16{0x2b2b}},
		{modelName: "Gray16", input: color.Gray16{0x593}, expected: color.Gray16{0x505}},
		{modelName: "Gray16", input: color.Gray16{0x1AAC}, expected: color.Gray16{0x1a1a}},
		{modelName: "Gray16", input: color.Gray16{0xE802}, expected: color.Gray16{0xe8e8}},
		{modelName: "Gray16", input: color.Gray16{0xB10}, expected: color.Gray16{0xb0b}},
		{modelName: "Gray16", input: color.Gray16{0xFC45}, expected: color.Gray16{0xfcfc}},
	}

	for idx, tt := range tests {
		t.Run(fmt.Sprintf("%d - %v to %v with model %s", idx, tt.input, tt.expected, tt.modelName), func(t *testing.T) {
			res := flatten(rgbaFromColor(tt.input))
			if !approxEq(res, tt.expected, 0xff) {
				t.Fatalf("expected %v, got %v", tt.expected, res)
			}
		})
	}
}

func approxEq(c1, c2 color.Gray16, th uint16) bool {
	// floats, otherwise we have to deal with over/underflows
	return math.Abs(float64(c1.Y)-float64(c2.Y)) < float64(th)
}
