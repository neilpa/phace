package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"neilpa.me/phace"
)

// Oultilnes produces a transparent image with rectangles where the
// faces should be.
func Outlines(photo *phace.Photo, faces []*phace.Face) *image.RGBA {
	// Need to create a mutable version of the image
	bounds := image.Rect(0, 0, photo.Width, photo.Height)
	dst := image.NewRGBA(bounds)
	fmt.Println("faces", len(faces))

	border := 10
	for _, f := range faces {
		// Convert to pixel coords, size is aligned to smaller dimension
		width, height := float64(bounds.Dx()), float64(bounds.Dy())

		// Trial and error suggests f.Size is the radius about the center
		// of the face. A box is close enough for validating.
		radius := int(f.Size * width) // TODO Round?
		if height < width {
			radius = int(f.Size * height)
		}
		center := makePoint(f.CenterX, f.CenterY, bounds, photo.Orientation)
		min := image.Pt(center.X - radius, center.Y - radius)
		max := image.Pt(center.X + radius, center.Y + radius)

		top := image.Rect(min.X-border, min.Y-border, max.X+border, min.Y)
		bot := image.Rect(min.X-border, max.Y, max.X+border, max.Y+border)
		left := image.Rect(min.X-border, min.Y, min.X, max.Y)
		right := image.Rect(max.X, min.Y, max.X+border, max.Y)

		draw.Draw(dst, top, blue, image.ZP, draw.Src)
		draw.Draw(dst, bot, green, image.ZP, draw.Src)
		draw.Draw(dst, left, red, image.ZP, draw.Src)
		draw.Draw(dst, right, gray, image.ZP, draw.Src)

		drawDot(dst, f.LeftEyeX, f.LeftEyeY, bounds, photo.Orientation, blue)
		drawDot(dst, f.RightEyeX, f.RightEyeY, bounds, photo.Orientation, red)
		drawDot(dst, f.MouthX, f.MouthY, bounds, photo.Orientation, green)
		drawDot(dst, f.CenterX, f.CenterY, bounds, photo.Orientation, black)
	}

	return dst
}

func drawDot(dst draw.Image, x, y float64, bounds image.Rectangle, orientation int, c image.Image) {
	pt := makePoint(x, y, bounds, orientation)
	sz := 15 // TODO Scale relative to size of face?
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
