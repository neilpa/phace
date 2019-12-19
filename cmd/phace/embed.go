package main

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"
	"strings"

	"neilpa.me/phace"
	"neilpa.me/phace/mwg-rs"

	"neilpa.me/go-jfif"
	"trimmer.io/go-xmp/models/xmp_base"
	"trimmer.io/go-xmp/models/xmp_tpg"
	"trimmer.io/go-xmp/xmp"
)

const (
	sigXMP         = "http://ns.adobe.com/xap/1.0/\x00"
	sigExtendedXMP = "http://ns.adobe.com/xmp/extension/\x00"
)

// EmbedFaces creates a copy of the photo's image with facial tags embeded
// as metadata.
//
// TODO Specifics and that this only works for JPEGs currently.
func EmbedFaces(s *phace.Session, p *phace.Photo, faces []*phace.Face) error {
	f, err := os.Open(s.ImagePath(p))
	if err != nil {
		return err
	}
	defer f.Close()

	if faces == nil {
		faces, err = p.Faces(s)
		if err != nil {
			return err
		}
		// TODO No Faces?
	}

	segments, err := jfif.DecodeSegments(f)
	if err != nil {
		return err
	}

	// TODO For now simply bail if there are existing XMP segments rather than
	// merging since this isn't the case for any of my photos
	var head, tail []jfif.Segment
	for _, seg := range segments {
		//fmt.Println(seg.Marker, len(seg.Data))
		if m := seg.Marker; m != jfif.APP1 {
			if m == jfif.SOI || (jfif.APP0 <= m && m >= jfif.APP15) {
				fmt.Println("Appending head", m)
				head = append(head, seg)
			} else {
				fmt.Println("Appending tail", m)
				tail = append(tail, seg)
			}
			continue
		}
		if strings.HasPrefix(string(seg.Data), sigXMP) {
			return fmt.Errorf("TODO: Existing XMP segment %s", p.Path)
		}
		if strings.HasPrefix(string(seg.Data), sigExtendedXMP) {
			return fmt.Errorf("TODO: Existing ExtendedXMP segment %s", p.Path)
		}
		fmt.Println("Appending end", seg.Marker)
		head = append(head, seg)
	}

	// TODO Grab the size from one of the segments above
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	config, _, err := image.DecodeConfig(f)
	if err != nil {
		return err
	}

	// TODO I've seen images with a preview at the end of them. after the main image
	// What this should do instead is only decode the metadata and then copy over
	// everything from the SOS and beyond

	// Create the new XMP segment
	var buf bytes.Buffer
	buf.WriteString(sigXMP)

	enc := xmp.NewEncoder(&buf)
	doc := xmp.NewDocument()
	doc.AddModel(makeRegions(config, faces))
	err = enc.Encode(doc)
	if err != nil {
		return err
	}
	fmt.Printf("segments %d head %d tail %d\n", len(segments), len(head), len(tail))
	faceSegment := jfif.Segment{jfif.APP1, buf.Bytes(), -1}
	head = append(head, faceSegment)
	fmt.Printf("segments %d head %d tail %d\n", len(segments), len(head), len(tail))

	// TODO Do any of the metadata sections track the total file size?
	path := filepath.Join("out", filepath.Base(p.Path))
	w, err := os.Create(path)
	if err != nil {
		return err
	}
	defer w.Close()

	// Write the SOI and APPn segments
	if err = writeSegments(w, head); err != nil {
		return err
	}
	// Write the remaining segments
	return writeSegments(w, tail)
}

func makeRegions(config image.Config, faces []*phace.Face) *mwgrs.Regions {
	regionList := make(mwgrs.RegionStructList, len(faces))
	for i, f := range faces {
		regionList[i] = mwgrs.RegionStruct{
			// TODO Get the name from a face group
			// TODO Does orientation matter?
			Type: mwgrs.TypeFace,
			Area: xmpbase.Area{
				X:    f.CenterX,
				Y:    f.CenterY,
				D:    f.Size,
				Unit: "normalized",
			},
		}
	}

	return &mwgrs.Regions{
		Regions: mwgrs.RegionInfo{
			AppliedToDimensions: xmptpg.Dimensions{
				H:    float32(config.Height),
				W:    float32(config.Width),
				Unit: "pixel",
			},
			RegionList: regionList,
		},
	}
}

func writeSegments(w io.Writer, segments []jfif.Segment) error {
	for _, seg := range segments {
		if err := jfif.EncodeSegment(w, seg); err != nil {
			return err
		}
	}
	return nil
}
