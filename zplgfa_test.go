package zplgfa

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"
	"testing"
)

type zplTest struct {
	Filename    string `json:"filename"`
	Zplstring   string `json:"zplstring"`
	Graphictype string `json:"graphictype"`
}

func Test_CompressASCII(t *testing.T) {
	const hexstr = "FFFFFFFF000000"
	buf := bytes.NewBuffer(make([]byte, 0, len(hexstr)))
	CompressASCII(buf, hexstr)
	if buf.String() != "NFL0" {
		t.Fatalf("CompressASCII failed")
	}
}

func Test_ConvertToZPL(t *testing.T) {
	f, err := os.Open("./tests/tests.json")
	if err != nil {
		t.Fatalf("error opening tests file: %v", err)
	}
	var zplTests []zplTest
	if err := json.NewDecoder(f).Decode(&zplTests); err != nil {
		t.Fatalf("error decoding json test file: %v", err)
	}
	f.Close()
	if len(zplTests) < 12 {
		t.Fatal("expected at least 12 tests to be run")
	}
	var graphicType GraphicType
	for i, testcase := range zplTests {
		t.Run(fmt.Sprintf("%d.%s", i, testcase.Filename), func(t *testing.T) {
			filename, zplstring, graphictype := testcase.Filename, testcase.Zplstring, testcase.Graphictype
			// open file
			file, err := os.Open(filename)
			if err != nil {
				t.Errorf("Warning: could not open the file \"%s\": %s\n", filename, err)
				return
			}
			defer file.Close()

			// load image head information
			config, format, err := image.DecodeConfig(file)
			if err != nil {
				t.Errorf("Warning: image not compatible, format: %s, config: %v, error: %s\n", format, config, err)
			}

			// reset file pointer to the beginning of the file
			file.Seek(0, 0)

			// load and decode image
			img, _, err := image.Decode(file)
			if err != nil {
				t.Errorf("Warning: could not decode the file, %s\n", err)
				return
			}

			// flatten image
			flat := FlattenImage(img)

			// convert image to zpl compatible type
			switch graphictype {
			case "ASCII":
				graphicType = ASCII
			case "Binary":
				graphicType = Binary
			case "CompressedASCII":
				graphicType = CompressedASCII
			default:
				graphicType = CompressedASCII
			}

			gfimg := ConvertToZPL(flat, graphicType)

			if graphictype == "Binary" {
				gfimg = base64.StdEncoding.EncodeToString([]byte(gfimg))
			} else {
				// remove whitespace - only for the test
				gfimg = strings.Replace(gfimg, " ", "", -1)
				gfimg = strings.Replace(gfimg, "\n", "", -1)
			}

			if gfimg != zplstring {
				t.Errorf("ConvertToZPL Test for file \"%s\" failed, wanted: \n%s\ngot: \n%s\n", filename, zplstring, gfimg)
			}
		})
	}
}
