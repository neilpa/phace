package main

import (
	"bytes"
	"fmt"
	"io"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"
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

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: phace <*.photoslibrary>")
		os.Exit(2)
	}
	s, err := phace.OpenSession(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	Dump(s)
}

// Dump prints paths to image files and raw face tag data.
func Dump(s *phace.Session) {
	photos, err := s.Photos()

	// TODO Use more cores
	for _, p := range photos {
		faces, err := p.Faces(s)
		if len(faces) == 0 {
			continue;
		}

		fmt.Printf("%s orientation=%d type=%d adj=%t\n", p.Path, p.Orientation, p.Type, p.HasAdjustments)
		for _, f := range faces {
			fmt.Printf("  %s center=(%f,%f) size=%f left=(%f,%f) right=(%f,%f), mouth=(%f,%f)\n",
				f.GroupUUID, f.CenterX, f.CenterY, f.Size, f.LeftEyeX, f.LeftEyeY, f.RightEyeX, f.RightEyeY, f.MouthX, f.MouthY)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		if err := EmbedFaces(s, p, faces); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		// Quick check if things are working...
		//if err := OutlineFaces(s, p, faces); err != nil {
		//	fmt.Fprintln(os.Stderr, err)
		//}
	}
	if err != nil {
		log.Fatal(err)
	}
}

const (
	sigXMP = "http://ns.adobe.com/xap/1.0/\x00"
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
				X: f.CenterX,
				Y: f.CenterY,
				D: f.Size,
				Unit: "normalized",
			},
		}
	}

	return &mwgrs.Regions{
		Regions: mwgrs.RegionInfo{
			AppliedToDimensions:xmptpg.Dimensions{
				H: float32(config.Height),
				W: float32(config.Width),
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

// OutlineFaces creates a new image, drawing a boarder around the faces.
// Dumps the resulting image in `out/` with the same basename as the
// original. Best way to check how things are actually working...
func OutlineFaces(s *phace.Session, p *phace.Photo, faces []*phace.Face) error {
	src, err := s.Image(p)
	if err != nil {
		return err
	}
	if faces == nil {
		faces, err = p.Faces(s)
		if err != nil {
			return err
		}
	}

	// Need to create a mutable version of the image
	r := src.Bounds()
	dst := image.NewRGBA(r)
	draw.Draw(dst, r, src, image.ZP, draw.Src)

	border := 10
	for _, f := range faces {
		// Trial and error suggests f.Size is a rough radius
		// about the center of the face. A box is close enough
		// for validating.
		min := makePoint(f.CenterX-f.Size, f.CenterY-f.Size, r, p.Orientation)
		max := makePoint(f.CenterX+f.Size, f.CenterY+f.Size, r, p.Orientation)

		top := image.Rect(min.X-border, min.Y-border, max.X+border, min.Y)
		bot := image.Rect(min.X-border, max.Y, max.X+border, max.Y+border)
		left := image.Rect(min.X-border, min.Y, min.X, max.Y)
		right := image.Rect(max.X, min.Y, max.X+border, max.Y)

		draw.Draw(dst, top, blue, image.ZP, draw.Src)
		draw.Draw(dst, bot, green, image.ZP, draw.Src)
		draw.Draw(dst, left, red, image.ZP, draw.Src)
		draw.Draw(dst, right, gray, image.ZP, draw.Src)

		drawDot(dst, f.LeftEyeX, f.LeftEyeY, r, p.Orientation, blue)
		drawDot(dst, f.RightEyeX, f.RightEyeY, r, p.Orientation, red)
		drawDot(dst, f.MouthX, f.MouthY, r, p.Orientation, green)
		drawDot(dst, f.CenterX, f.CenterY, r, p.Orientation, black)
	}

	// Dump the files on disk for inspection
	err = os.MkdirAll("out", 0755)
	if err != nil {
		return err
	}
	w, err := os.Create(filepath.Join("out", filepath.Base(p.Path)))
	if err != nil {
		return err
	}
	return jpeg.Encode(w, dst, nil)
}

func drawDot(dst draw.Image, x, y float64, bounds image.Rectangle, orientation int, c image.Image) {
	pt := makePoint(x, y, bounds, orientation)
	sz := 15
	dot := image.Rect(pt.X-sz, pt.Y-sz, pt.X+sz, pt.Y+sz)
	draw.Draw(dst, dot, c, image.ZP, draw.Src)
}

func makePoint(x, y float64, r image.Rectangle, orientation int) image.Point {
	dx, dy := float64(r.Dx()), float64(r.Dy())
	switch orientation {
	case 1: // normal but not sure why the y axis is flipped
		y = 1 - y
	case 3: // upside down too
		x = 1 - x
	case 6: // portrait
		x, y = 1-y, 1-x
	case 8: // only 1 example with a face
		x, y = y, x
	default:
		fmt.Println("unrecognized orientation:", orientation)
	}
	return image.Pt(int(x*dx), int(y*dy))
}

var (
	red   = &image.Uniform{color.RGBA{255, 0, 0, 255}}
	green = &image.Uniform{color.RGBA{0, 255, 0, 255}}
	blue  = &image.Uniform{color.RGBA{0, 0, 255, 255}}
	gray  = &image.Uniform{color.RGBA{100, 100, 100, 255}}
	black = &image.Uniform{color.RGBA{0, 0, 0, 255}}
)
