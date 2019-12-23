package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"

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
	if err := os.MkdirAll(*out, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
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

	var wg sync.WaitGroup
	// Limit the number of simultaneously open files to avoid ulimit issues
	sem := make(chan struct{}, runtime.GOMAXPROCS(0)+10)

	for _, p := range photos {
		faces, err := p.Faces(s)
		if len(faces) == 0 {
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
			}
			continue
		}

		wg.Add(1)
		go func(p *phace.Photo, faces []*phace.Face) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			var err error

			debug("%s orientation=%d type=%d adj=%t", p.Path, p.Orientation, p.Type, p.HasAdjustments)
			for _, f := range faces {
				debug("%s group=%s center=(%f,%f) size=%f left=(%f,%f) right=(%f,%f), mouth=(%f,%f)", p.Path,
					f.GroupUUID, f.CenterX, f.CenterY, f.Size, f.LeftEyeX, f.LeftEyeY, f.RightEyeX, f.RightEyeY, f.MouthX, f.MouthY)
			}

			switch {
			case *embed:
				err = EmbedFaces(s, p, faces, *out)
				if err == nil {
					fmt.Printf("embedded %d faces in %s\n", len(faces), p.Path)
				}
			case *fdraw:
				err = OutlineFaces(s, p, faces, *out)
				debug("outlined %s", p.Path)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
			}
		}(p, faces)
	}
	wg.Wait()
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
