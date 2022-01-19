package zplgfa

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"strings"
)

// GraphicType is a type to select the graphic format
type GraphicType int

func (gt GraphicType) String() string {
	if gt == Binary {
		return "B"
	}
	return "A"
}

const (
	// ASCII graphic type using only hex characters (0-9A-F)
	ASCII GraphicType = iota
	// Binary saving the same data as binary
	Binary
	// CompressedASCII compresses the hex data via RLE
	CompressedASCII
)

// ConvertToZPL is just a wrapper for ConvertToGraphicField which also includes the ZPL
// starting code ^XA and ending code ^XZ, as well as a Field Separator and Field Origin.
func ConvertToZPL(img image.Image, graphicType GraphicType) string {
	return fmt.Sprintf("^XA,^FS\n^FO0,0\n%s^FS,^XZ\n", ConvertToGraphicField(img, graphicType))
}

var (
	shortWhite = flatten(rgbaFromColor(color.White))
	shortBlack = flatten(rgbaFromColor(color.Black))
)

func whiteish(v uint32) bool {
	// colors represented in u32 are in range [0, 0xffff]
	// see type image.Color interface docs
	return v > 0xff00
}

func blackish(v uint32) bool {
	// colors represented in u32 are in range [0, 0xffff]
	// see type image.Color interface docs
	return v < 0x00ff
}

type rgba struct{ r, g, b, a uint32 }

func (c rgba) RGBA() (r, g, b, a uint32) {
	return c.r, c.g, c.b, c.a
}

func rgbaFromColor(c color.Color) rgba {
	r, g, b, a := c.RGBA()
	return rgba{r, g, b, a}
}

func shortcircuit(input rgba) (color.Gray16, bool) {
	r, g, b, a := input.RGBA()
	if whiteish(r) && whiteish(g) && whiteish(b) && whiteish(a) {
		return shortWhite, true
	}
	if blackish(r) && blackish(g) && blackish(b) && blackish(a) {
		return shortBlack, true
	}
	return color.Gray16{}, false
}

// FlattenImage optimizes an image for the converting process
// Not really needed as ConvertToGraphicField already does this internally
// to avoid looping through image (and doing image.At calls) twice
func FlattenImage(source image.Image) *image.Gray16 {
	size := source.Bounds().Size()
	target := image.NewGray16(source.Bounds())
	for y := 0; y < size.Y; y++ {
		for x := 0; x < size.X; x++ {
			p := source.At(x, y)
			rgba := rgbaFromColor(p)
			flat, ok := shortcircuit(rgba)
			if !ok {
				flat = flatten(rgba)
			}
			target.SetGray16(x, y, flat)
		}
	}
	return target
}

// adapted from color.Gray16Model.Convert
func gray16Model(r, g, b uint32) color.Gray16 {
	// These coefficients (the fractions 0.299, 0.587 and 0.114) are the same
	// as those given by the JFIF specification and used by func RGBToYCbCr in
	// ycbcr.go.
	//
	// Note that 19595 + 38470 + 7471 equals 65536.
	y := (19595*r + 38470*g + 7471*b + 1<<15) >> 16
	return color.Gray16{uint16(y)}
}

func flatten(input rgba) color.Gray16 {
	r, g, b, a := input.RGBA()
	alpha := float32(a) / 0xffff
	val := 0xffff - uint32((float32(color.White.Y) * alpha))
	conv := func(c uint32) uint32 {
		return val | uint32(float32(c)*alpha)
	}
	return gray16Model(conv(r), conv(g), conv(b))
}

func writeRepeatCode(dst io.Writer, repeatCount int, char rune) int {
	n := 0
	if repeatCount > 419 {
		repeatCount -= 419
		n += writeRepeatCode(dst, repeatCount, char)
		repeatCount = 419
	}

	high := repeatCount / 20
	low := repeatCount % 20

	const lowString = " GHIJKLMNOPQRSTUVWXY"
	const highString = " ghijklmnopqrstuvwxyz"

	if high > 0 {
		n += mustWrite(dst, []byte{highString[high]})
	}
	if low > 0 {
		n += mustWrite(dst, []byte{lowString[low]})
	}
	n += mustWrite(dst, []byte(string(char)))
	return n
}

// CompressASCII compresses the ASCII data of a ZPL Graphic Field using RLE
func CompressASCII(dst io.Writer, in string) {
	var lastChar rune
	var lastCharSince int
	haveWritten := false

	update := func(i int) {
		if i-lastCharSince > 4 {
			if n := writeRepeatCode(dst, i-lastCharSince, lastChar); n > 0 {
				haveWritten = true
			}
			return
		}
		for j := 0; j < i-lastCharSince; j++ {
			if n := mustWrite(dst, []byte(string(lastChar))); n > 0 {
				haveWritten = true
			}
		}
	}

	for i, curChar := range in {
		if lastChar == curChar {
			continue
		}
		update(i)
		lastChar = curChar
		lastCharSince = i
	}
	if lastCharSince == 0 {
		switch lastChar {
		case '0':
			mustWrite(dst, []byte(","))
			return
		case 'F':
			mustWrite(dst, []byte("!"))
			return
		}
	}
	update(len(in))

	if !haveWritten {
		writeRepeatCode(dst, len(in), lastChar)
	}
}

func mustWrite(dst io.Writer, s []byte) int {
	n, err := dst.Write(s)
	if err != nil {
		panic(err)
	}
	return n
}

// ConvertToGraphicField converts an image.Image picture to a ZPL compatible Graphic Field.
// The ZPL ^GF (Graphic Field) supports various data formats, this package supports the
// normal ASCII encoded, as well as a RLE compressed ASCII format. It also supports the
// Binary Graphic Field format. The encoding can be chosen by the second argument.
func ConvertToGraphicField(source image.Image, graphicType GraphicType) string {
	size := source.Bounds().Size()
	width := size.X / 8
	height := size.Y
	if size.Y%8 != 0 {
		width = width + 1
	}

	dst := bytes.NewBuffer(make([]byte, 0, 8*1024))
	var compressionBuf *bytes.Buffer
	readyCompressionBuf := func(hexstr string) {
		if compressionBuf == nil {
			compressionBuf = bytes.NewBuffer(make([]byte, 0, len(hexstr)))
			return
		}
		compressionBuf.Reset()
		if len(hexstr) > compressionBuf.Cap() {
			compressionBuf.Grow(len(hexstr) - compressionBuf.Cap())
		}
	}

	// adapted from: https://go-review.googlesource.com/c/go/+/72370
	pxRGBA := func(x, y int) (r, g, b, a uint32) { return source.At(x, y).RGBA() }
	// Fast paths for special cases to avoid excessive use of the color.Color
	// interface which escapes to the heap but need to be discovered for
	// each pixel on r. See also https://golang.org/issues/15759.
	switch src0 := source.(type) {
	case *image.RGBA:
		pxRGBA = func(x, y int) (r, g, b, a uint32) { return src0.RGBAAt(x, y).RGBA() }
	case *image.NRGBA:
		pxRGBA = func(x, y int) (r, g, b, a uint32) { return src0.NRGBAAt(x, y).RGBA() }
	case *image.RGBA64:
		pxRGBA = func(x, y int) (r, g, b, a uint32) { return src0.RGBA64At(x, y).RGBA() }
	case *image.NRGBA64:
		pxRGBA = func(x, y int) (r, g, b, a uint32) { return src0.NRGBA64At(x, y).RGBA() }
	case *image.YCbCr:
		pxRGBA = func(x, y int) (r, g, b, a uint32) { return src0.YCbCrAt(x, y).RGBA() }
	}

	var lastLine string
	for y := 0; y < size.Y; y++ {
		line := make([]uint8, width)
		lineIndex := 0
		index := uint8(0)
		currentByte := line[lineIndex]
		for x := 0; x < size.X; x++ {
			index = index + 1
			r, g, b, a := pxRGBA(x, y)
			rgba := rgba{r, g, b, a}
			lum, ok := shortcircuit(rgba)
			if !ok {
				lum = flatten(rgba)
			}
			if lum.Y < math.MaxUint16/2 {
				currentByte = currentByte | (1 << (8 - index))
			}
			if index >= 8 {
				line[lineIndex] = currentByte
				lineIndex++
				if lineIndex < len(line) {
					currentByte = line[lineIndex]
				}
				index = 0
			}
		}

		hexstr := strings.ToUpper(hex.EncodeToString(line))

		switch graphicType {
		case ASCII:
			mustWrite(dst, []byte(hexstr+"\n"))
		case CompressedASCII:
			readyCompressionBuf(hexstr)
			CompressASCII(compressionBuf, hexstr)
			if lastLine == compressionBuf.String() {
				mustWrite(dst, []byte(":"))
			} else {
				mustWrite(dst, compressionBuf.Bytes())
			}
			lastLine = compressionBuf.String()
		case Binary:
			mustWrite(dst, []byte(line))
		}
	}

	return fmt.Sprintf("^GF%s,%d,%d,%d,\n", graphicType.String(), dst.Len(), width*height, width) + dst.String()
}
