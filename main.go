package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/llgcode/draw2d/draw2dimg"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: phace <*.photoslibrary>")
		os.Exit(2)
	}
	s, err := OpenSession(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	Dump(s)
}

// Session wraps the photos library folder and sqlite connections.
type Session struct {
	// Path to the root *.photoslibrary folder.
	Path string
	// LibraryDB is the sqlite connection with photo data.
	LibraryDB *sqlx.DB
	// PersonDB is the sqlite connectin with face data.
	PersonDB *sqlx.DB
}

// OpenSession connects to the embedded sqlite databases.
func OpenSession(path string) (*Session, error) {
	dbPath := filepath.Join(path, "database")
	libraryDB, err := openDB(filepath.Join(dbPath, "Library.apdb"))
	if err != nil {
		return nil, err
	}
	personDB, err := openDB(filepath.Join(dbPath, "Person.db"))
	if err != nil {
		return nil, err
	}
	return &Session{path, libraryDB, personDB}, nil
}

// Image opens the master image file in the library.
func (s *Session) Image(p *Photo) (image.Image, error) {
	f, err := os.Open(filepath.Join(s.Path, "Masters", p.Path))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m, _, err := image.Decode(f)
	return m, err
}

// Dump prints paths to image files and raw face tag data.
func Dump(s *Session) {
	photos, err := s.Photos()

	for _, p := range photos {
		fmt.Println(p.Path)
		faces, err := p.Faces(s)
		for _, f := range faces {
			fmt.Printf("  %s center=(%f,%f) size=%f left=(%f,%f) right=(%f,%f), mouth=(%f,%f)\n",
				f.GroupUUID, f.CenterX, f.CenterY, f.Size, f.LeftEyeX, f.LeftEyeY, f.RightEyeX, f.RightEyeY, f.MouthX, f.MouthY)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

        // Quick check if things are working...
		if len(faces) > 0 {
            if err := OutlineFaces(s, p); err != nil {
			    fmt.Fprintln(os.Stderr, err)
            }
        }
	}
	if err != nil {
		log.Fatal(err)
	}
}

// Photos gets all the photo records in the library.
func (s *Session) Photos() ([]*Photo, error) {
	photos := make([]*Photo, 0)
	err := s.LibraryDB.Select(&photos, `
        SELECT v.uuid, v.masterUuid, m.fingerprint, m.imagePath
        FROM RKVersion v
        JOIN RKMaster m ON m.uuid = v.masterUuid
    `)
	return photos, err
}

// Faces gets all the face records in the library.
func (s *Session) Faces() ([]*Face, error) {
	faces := make([]*Face, 0)
	err := s.PersonDB.Select(&faces, `
        SELECT f.uuid, fg.uuid AS groupId, f.imageId, f.centerX, f.centerY, f.size
            f.
        SELECT f.uuid, fg.uuid AS groupId, f.imageId, f.centerX, f.centerY, f.size
        FROM RKFace f
        JOIN RKFaceGroupFace fgf ON fgf.faceId = f.modelId
        JOIN RKFaceGroup fg ON fg.modelId = fgf.faceGroupId
    `)
	return faces, err
}

// FaceGroups gets all the group records in the library.
func (s *Session) FaceGroups() ([]*Face, error) {
	return nil, fmt.Errorf("todo")
}

// Photo record in the library, references an image on disk.
type Photo struct {
	UUID        string `db:"uuid"`
	MasterUUID  string `db:"masterUuid"`
	Fingerprint string `db:"fingerprint"`
	// Path to the image on disk, relative to the library root.
	Path string `db:"imagePath"`

	Height int `db:"masterHeight"`
	Width  int `db:"masterWidth"`
}

// Faces gets the recognized faces in the photo.
func (p *Photo) Faces(s *Session) ([]*Face, error) {
	faces := make([]*Face, 0)
	err := s.PersonDB.Select(&faces, `
        SELECT f.uuid, fg.uuid AS groupId, f.imageId, f.centerX, f.centerY, f.size,
            f.leftEyeX, f.leftEyeY, f.rightEyeX, f.rightEyeY, f.mouthX, f.mouthY,
            f.hasSmile, f.isBlurred, f.isLeftEyeClosed, f.isRightEyeClosed
        FROM RKFace f
        JOIN RKFaceGroupFace fgf ON fgf.faceId = f.modelId
        JOIN RKFaceGroup fg ON fg.modelId = fgf.faceGroupId
        WHERE imageId = ?
    `, p.UUID)
	return faces, err
}

// Face record in the library, a recognized face in a single photo.
type Face struct {
	UUID      string `db:"uuid"`
	GroupUUID string `db:"groupId"`
	ImageUUID string `db:"imageId"`

	CenterX        float64 `db:"centerX"`
	CenterY        float64 `db:"centerY"`
	Size           float64 `db:"size"`

	LeftEyeX       float64 `db:"leftEyeX"`
	LeftEyeY       float64 `db:"leftEyeY"`
	RightEyeX      float64 `db:"rightEyeX"`
	RightEyeY      float64 `db:"rightEyeY"`
	MouthX         float64 `db:"mouthX"`
	MouthY         float64 `db:"mouthY"`

	HasSmile       bool    `db:"hasSmile"`
	IsBlurred      bool    `db:"isBlurred"`
	IsLeftEyeClosed  bool    `db:"isLeftEyeClosed"`
	IsRightEyeClosed bool    `db:"isRightEyeClosed"`
}

// Faces gets other instances of faces that belong to the same group.
func (f Face) FaceGroup(s *Session) (*FaceGroup, error) {
	return nil, fmt.Errorf("todo")
}

// Photo gets the entity contaning this face.
func (f Face) Photo(s *Session) (*Photo, error) {
	return nil, fmt.Errorf("todo")
}

// FaceGroup record in the library, a collection of recognized faces that
// _should_ belong to the same individual.
type FaceGroup struct {
	UUID     string `db:"uuid"`
	PersonId int    `db:"personId"`
}

// openDB creates a sqlite connection and issues a test query to ensure
// the database isn't locked.
func openDB(path string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("phace: %s: %v", path, err)
	}
	return db, err
}

// Outilne creates a new image, drawing a boarder around the faces. Dumps the
// resulting image in `out/` with the same basename as the original. Best way
// to check how things are actually working...
func OutlineFaces(s *Session, p *Photo) error {
    src, err := s.Image(p)
    if err != nil {
        return err
    }
    faces, err := p.Faces(s)
    if err != nil {
        return err
    }

    // Need to create a mutable version of the image
	r := src.Bounds()
	dst := image.NewRGBA(r)
	draw.Draw(dst, r, src, image.ZP, draw.Src)

	gc := draw2dimg.NewGraphicContext(dst)
	gc.SetStrokeColor(color.RGBA{0xff, 0x00, 0x00, 0xff})
	gc.SetLineWidth(10)
	gc.BeginPath()

	dx, dy := float64(r.Dx()), float64(r.Dy())
	for _, f := range faces {
        // Trial and error suggests f.Size is a rough radius
        // about the center of the face. A box is close enough
        // for validating.
		minX := (f.CenterX - f.Size) * dx
		maxX := (f.CenterX + f.Size) * dx
		minY := (f.CenterY - f.Size) * dy
		maxY := (f.CenterY + f.Size) * dy

		gc.MoveTo(minX, minY)
		gc.LineTo(minX, maxY)
		gc.LineTo(maxX, maxY)
		gc.LineTo(maxX, minY)
		gc.LineTo(minX, minY)
	}
	gc.Close()
	gc.Stroke()

    // Dump the files on disk for inspection
	w, err := os.Create(filepath.Join("out", filepath.Base(p.Path)))
	if err != nil {
		return err
	}
	return jpeg.Encode(w, dst, nil)
}
