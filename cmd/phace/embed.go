package main

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"

	"neilpa.me/phace"
	"neilpa.me/phace/mwg-rs"

	"neilpa.me/go-jfif"
	"neilpa.me/go-x/io"
	"trimmer.io/go-xmp/models/xmp_base"
	//"trimmer.io/go-xmp/models/xmp_tpg"
	"trimmer.io/go-xmp/xmp"
)

// EmbedFaces creates a copy of the photo's image with facial tags embeded
// as metadata.
//
// TODO Specifics and that this only works for JPEGs currently.
func EmbedFaces(s *phace.Session, p *phace.Photo, faces []*phace.Face, dir string) error {
	path := s.ImagePath(p)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// TODO Add a new XMP segment if one does not exist
	//		Otherwise create an XMP extended segment with the new data
	segs, err := jfif.ScanSegments(f)
	if err != nil {
		return err
	}

	var xmpSeg *jfif.Segment
	var imgStart = int64(-1)
	var doc *xmp.Document
	for _, seg := range segs {
		if m := seg.Marker; m != jfif.APP1 {
			if imgStart < 0 && m != jfif.SOI && (m < jfif.APP0 || m > jfif.APP15) {
				// First non-metadata segment related to the image
				imgStart = seg.Offset
			}
			continue
		}
		sig, payload, _ := seg.AppPayload()
		if sig != jfif.SigXMP {
			continue
		}
		if xmpSeg != nil {
			return fmt.Errorf("%s: multiple XMP segments", path)
		}
		doc = &xmp.Document{}
		if err = xmp.Unmarshal(payload, doc); err != nil {
			return err
		}
		xmpSeg, doc = &seg, doc
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

	newpath := filepath.Join(dir, filepath.Base(p.Path))
	w, err := os.Create(newpath)
	if err != nil {
		return err
	}
	defer w.Close()

	// TODO Switch to the jfif.File interface

	// Some JPEG files have multiple images (e.g. a preview as a trailer
	// image). To handle this we copy from the original, inserting the
	// new/modified XMP data at the appropriate poinit.
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	if xmpSeg != nil {
		// Write the partial header and skip the original xmp segment
		_, err = io.Copy(w, &io.LimitedReader{f, xmpSeg.Offset})
		if err != nil {
			return err
		}
		_, err = f.Seek(int64(xmpSeg.Size), io.SeekCurrent)
	} else {
		// Write until the first image data segment
		if imgStart < 0 {
			return fmt.Errorf("%s: missing image segment", path)
		}
		_, err = io.Copy(w, &io.LimitedReader{f, imgStart})
	}
	if err != nil {
		return err
	}

	// Write the new xmp segment and remainder of the original file
	if err = encodeFaces(w, doc, config, faces); err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}

// encodeFaces writes a new XMP segment to w
func encodeFaces(w io.Writer, doc *xmp.Document, config image.Config, faces []*phace.Face) error {
	// Create the new XMP segment
	var buf bytes.Buffer
	buf.WriteString(jfif.SigXMP)

	if doc == nil {
		doc = xmp.NewDocument()
	}
	model, err := doc.MakeModel(mwgrs.NsMwgRs)
	if err != nil {
		return err
	}
	regionsModel := model.(*mwgrs.Regions)
	// TODO Validate that dimensions on existing data...

	regionList := makeRegionList(config, faces) // TODO This is lame
	regionsModel.Regions.RegionList = append(regionsModel.Regions.RegionList, regionList...)

	enc := xmp.NewEncoder(&buf)
	enc.Indent("", "  ")
	err = enc.Encode(doc)
	if err != nil {
		return err
	}

	seg := jfif.NewSegment(jfif.APP1, buf.Bytes())
	return seg.Write(w)
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
