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
	//"trimmer.io/go-xmp/models/xmp_tpg"
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
func EmbedFaces(s *phace.Session, p *phace.Photo, faces []*phace.Face, dir string) error {
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

	var head, tail []jfif.Segment
	var doc *xmp.Document
	for _, seg := range segments {
		if m := seg.Marker; m != jfif.APP1 {
			if m == jfif.SOI || (jfif.APP0 <= m && m >= jfif.APP15) {
				head = append(head, seg)
			} else {
				tail = append(tail, seg)
			}
			continue
		}

		if strings.HasPrefix(string(seg.Data), sigXMP) {
			// Drop the existing XMP
			doc = &xmp.Document{}
			err = xmp.Unmarshal(seg.Data[len(sigXMP):], doc)
			if err != nil {
				return err
			}
		} else if strings.HasPrefix(string(seg.Data), sigExtendedXMP) {
			// TODO: Do I need to worry about this?
			return fmt.Errorf("TODO: Existing ExtendedXMP segment %s", p.Path)
		} else {
			// Save this segment
			head = append(head, seg)
		}
	}

	// TODO Grab the image size while decoding segments...
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
	// everything from the SOS and beyond.

	// Create the new XMP segment
	var buf bytes.Buffer
	buf.WriteString(sigXMP)

	if doc == nil {
		// TODO Skip for now
		return fmt.Errorf("%s: Skipping", p.Path)
		doc = xmp.NewDocument()
		// TODO Validate that dimensions on existing data...
	}
	model, err := doc.MakeModel(mwgrs.NsMwgRs)
	if err != nil {
		return err
	}
	regionsModel := model.(*mwgrs.Regions)

	regionList := makeRegionList(config, faces) // TODO This is lame
	regionsModel.Regions.RegionList = append(regionsModel.Regions.RegionList, regionList...)

	enc := xmp.NewEncoder(&buf)
	enc.Indent("", "  ")
	err = enc.Encode(doc)
	if err != nil {
		return err
	}
	head = append(head, jfif.Segment{jfif.APP1, buf.Bytes(), -1})

	// TODO Do any of the metadata sections track the total file size?
	path := filepath.Join(dir, filepath.Base(p.Path))
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

func makeRegionList(config image.Config, faces []*phace.Face) mwgrs.RegionStructList {
	regionList := make(mwgrs.RegionStructList, len(faces))
	for i, f := range faces {
		regionList[i] = mwgrs.RegionStruct{
			// TODO Get the name from a face group
			// TODO Does orientation matter 
			//	Yes - This needs similar conversion to draw so that they are in
			//		  the physical orientation of the image
			Type: mwgrs.TypeFace,
			Area: xmpbase.Area{
				X:    f.CenterX,
				Y:    f.CenterY,
				D:    2 * f.Size,
				Unit: "normalized",
			},
		}
	}

	return regionList

	//return &mwgrs.Regions{
	//	Regions: mwgrs.RegionInfo{
	//		AppliedToDimensions: xmptpg.Dimensions{
	//			H:    float32(config.Height),
	//			W:    float32(config.Width),
	//			Unit: "pixel",
	//		},
	//		RegionList: regionList,
	//	},
	//}
}

func writeSegments(w io.Writer, segments []jfif.Segment) error {
	for _, seg := range segments {
		if err := jfif.EncodeSegment(w, seg); err != nil {
			return err
		}
	}
	return nil
}
