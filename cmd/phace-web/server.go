package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"

	"neilpa.me/phace"
)

type server struct {
	session *phace.Session
	photos  []*phace.Photo

	// pages is the list of photo pages by date
	pages []*photoPage
	// index is used to lookup pages by photo UUID
	index map[string]*photoPage
}

// newServer creates a web server around a phace.Session at the
// provided photoslibrary path.
func newServer(photoslibrary string) (*server, error) {
	session, err := phace.OpenSession(photoslibrary)
	if err != nil {
		return nil, err
	}
	// TODO The photos method could embed the on-disk path?
	photos, err := session.Photos()
	if err != nil {
		return nil, err
	}
	sort.Slice(photos, func(i, j int) bool {
		return photos[i].ImageDate < photos[j].ImageDate
	})

	// Build the pages and an index
	pages := make([]*photoPage, len(photos))
	index := make(map[string]*photoPage, len(photos))
	for i, p := range photos {
		page := &photoPage{ImageUrl: imageUrl(session, p)}
		if i > 0 {
			page.Prev = photoUrl(photos[i-1])
		}
		if i < len(photos)-1 {
			page.Next = photoUrl(photos[i+1])
		}
		pages[i] = page
		index[p.UUID] = page
	}

	return &server{session, photos, pages, index}, nil
}

func (s *server) rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)

	if r.URL.Path != "/" {
		http.Error(w, "404 not found", 404)
		return
	}
	// TODO: Use a template as well with a better UX
	fmt.Fprintln(w, "<html><ul>")
	for _, p := range s.photos {
		fmt.Fprintf(w, "<li><a href='%s'>%s</a></li>", photoUrl(p), p.ImageDate)
	}
	fmt.Fprintln(w, `</ul></html>`)
}

func (s *server) photosHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)

	uuid := r.URL.Path
	page, ok := s.index[uuid]
	if !ok {
		http.Error(w, "404 not found", 404)
		return
	}
	if err := photoHtml.Execute(w, page); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func photoUrl(photo *phace.Photo) string {
	return "/photos/" + url.PathEscape(photo.UUID)
}

func imageUrl(session *phace.Session, photo *phace.Photo) string {
	return "/images/" + session.MasterPath(photo)
}
