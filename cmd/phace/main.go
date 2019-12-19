package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"neilpa.me/phace"
)

var (
	dbg   = flag.Bool("dbg", false, "print debug spew")
	fdraw = flag.Bool("draw", false, "draw outlines around faces")
	embed = flag.Bool("embed", true, "embed XMP face tags")
	out   = flag.String("out", "out", "output directory")
	// TODO sidecar file, svg/html wrapper, overwrite safety
)

func main() {
	flag.Parse()
	if flag.NArg() == 0 || *out == "" {
		fmt.Fprintf(os.Stderr, "usage: %s [OPTIONS] <*.photoslibrary>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(2)
	}
	if *embed && *fdraw {
		fmt.Fprintf(os.Stderr, "%s: TODO: -embed and -draw mutually exclusive\n", os.Args[0])
		os.Exit(2)
	}

	s, err := phace.OpenSession(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	photos, err := s.Photos()
	if err != nil {
		log.Fatal(err)
	}

	// TODO Use more cores
	for _, p := range photos {
		faces, err := p.Faces(s)
		if len(faces) == 0 {
			continue
		}

		debug("%s orientation=%d type=%d adj=%t", p.Path, p.Orientation, p.Type, p.HasAdjustments)
		for _, f := range faces {
			debug("%s group=%s center=(%f,%f) size=%f left=(%f,%f) right=(%f,%f), mouth=(%f,%f)", p.Path,
				f.GroupUUID, f.CenterX, f.CenterY, f.Size, f.LeftEyeX, f.LeftEyeY, f.RightEyeX, f.RightEyeY, f.MouthX, f.MouthY)
		}

		switch {
		case *embed:
			err = EmbedFaces(s, p, faces)
		case *fdraw:
			err = OutlineFaces(s, p, faces)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s", os.Args[0], err)
		}
	}
}

func debug(format string, head interface{}, tail ...interface{}) {
	if !*dbg {
		return
	}
	format = os.Args[0] + ": " + format + "\n"
	args := make([]interface{}, 1, len(tail)+1)
	args[0] = head
	args = append(args, tail...)
	fmt.Fprintf(os.Stderr, format, args...)
}
