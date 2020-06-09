// facetag adds XMP face tag(s) to the given image.
//
//	facetag -x int -y int -d diameter -n name IMAGE.jpg
//
//	x,y are the center of the face in pixesl
//	d is the diamter of the face in pixesl
//  n is the name of the person being tagged
//
// TODO: How to add multiple faces at a time?
//	facetag [options] face x y d [name] pet x y d [name] ...
//
//	facetag [global-opts] file.jpg [tags]
//		where tags look like <face|pet|barcode|focus> -x -y -d -n
//		and can be specified multiple times
//
// That's a good starting point
//		and then I can design the crud
//
// What about listing/deleting embeded faces
//
// Maybe starting with a sidecare file is the way to go
package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	dbg = flag.Bool("dbg", false, "print debug spew")
)

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "usage: %s [OPTIONS] <INPUT>...\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(2)
	}

	fmt.Println(flag.Args())
	//for _, arg := range flag.Args() {
	//	debug("%s", arg)
	//}
}

func debug(format string, head interface{}, tail ...interface{}) {
	if !*dbg {
		return
	}
	format = os.Args[0] + ": " + format + "\n"
	args := make([]interface{}, 1, len(tail) + 1)
	args[0] = head
	args = append(args, tail...)
	fmt.Fprintf(os.Stderr, format, args...)
}
