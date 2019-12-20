package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"path/filepath"

	"neilpa.me/phace"
)

// OutlineFaces creates a new image, drawing a boarder around the faces.
// Dumps the resulting image in `out/` with the same basename as the
// original. Best way to check how things are actually working...
func OutlineFaces(s *phace.Session, p *phace.Photo, faces []*phace.Face, dir string) error {
	src, err := s.Image(p)
	if err != nil {
		return err
	}
	if faces == nil {
		faces, err = p.Faces(s)
		if err != nil {
			return err
			fmt.Fprintf(os.Stderr, "%s: %s", os.Args[0], err)
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
	w, err := os.Create(filepath.Join(dir, filepath.Base(p.Path)))
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
