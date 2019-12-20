package mwgrs_test

import (
	//"fmt"
	"io/ioutil"
	"os"
	"testing"

	"neilpa.me/phace/mwg-rs"

	"trimmer.io/go-xmp/xmp"
	//_ "trimmer.io/go-xmp/models"
)

func TestModelUnmarshal(t *testing.T) {
	tests := []string{
		"testdata/apple-extensions.xmp",
		"testdata/wikipedia-example.xmp",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			buf, err := ioutil.ReadFile(tt)
			if err != nil {
				t.Fatal(err)
			}

			doc := &xmp.Document{}
			err = xmp.Unmarshal(buf, doc)
			if err != nil {
				t.Fatal(err)
			}

			model := doc.FindModel(mwgrs.NsMwgRs)
			//fmt.Printf("%#v\n", model)
			_ = model

			enc := xmp.NewEncoder(os.Stdout)
			enc.SetFlags(xmp.Xpacket | xmp.Xpadding)
			enc.SetMaxSize(int64(len(buf)))
			enc.Indent("", "  ")
			err = enc.Encode(doc)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestModelMarshal(t *testing.T) { // TODO
}
