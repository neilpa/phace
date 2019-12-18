package phace

import (
	"fmt"
	"image"
	"os"
	"path/filepath"

	// TODO Should callers be responsible for this?
	_ "image/jpeg"
	_ "image/png"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)


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

// ImagePath returns the on-disk path to the master image.
func (s *Session) ImagePath(p *Photo) string {
	return filepath.Join(s.Path, "Masters", p.Path)
}

// Image opens the master image file in the library.
func (s *Session) Image(p *Photo) (image.Image, error) {
	f, err := os.Open(s.ImagePath(p))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m, _, err := image.Decode(f)
	return m, err
}

// Photos gets all the photo records in the library.
func (s *Session) Photos() ([]*Photo, error) {
	photos := make([]*Photo, 0)
	err := s.LibraryDB.Select(&photos, `
        SELECT v.uuid, v.masterUuid, m.fingerprint, m.imagePath, v.orientation, v.type, v.hasAdjustments
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
        FROM RKFace f
        JOIN RKFaceGroupFace fgf ON fgf.faceId = f.modelId
        JOIN RKFaceGroup fg ON fg.modelId = fgf.faceGroupId
    `)
	return faces, err
}

// FaceGroups gets all the group records in the library.
func (s *Session) FaceGroups() ([]*FaceGroup, error) {
	return nil, fmt.Errorf("todo")
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

