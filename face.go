package main

import (
	"fmt"
)

// Face record in the library, a recognized face in a single photo.
type Face struct {
	UUID      string `db:"uuid"`
	GroupUUID string `db:"groupId"`
	ImageUUID string `db:"imageId"`

	CenterX float64 `db:"centerX"`
	CenterY float64 `db:"centerY"`
	Size    float64 `db:"size"`

	LeftEyeX  float64 `db:"leftEyeX"`
	LeftEyeY  float64 `db:"leftEyeY"`
	RightEyeX float64 `db:"rightEyeX"`
	RightEyeY float64 `db:"rightEyeY"`
	MouthX    float64 `db:"mouthX"`
	MouthY    float64 `db:"mouthY"`

	HasSmile         bool `db:"hasSmile"`
	IsBlurred        bool `db:"isBlurred"`
	IsLeftEyeClosed  bool `db:"isLeftEyeClosed"`
	IsRightEyeClosed bool `db:"isRightEyeClosed"`
}

// FaceGroup record in the library, a collection of recognized faces that
// _should_ belong to the same individual.
type FaceGroup struct {
	UUID     string `db:"uuid"`
	PersonId int    `db:"personId"`
}

// Faces gets other instances of faces that belong to the same group.
func (f Face) FaceGroup(s *Session) (*FaceGroup, error) {
	return nil, fmt.Errorf("todo")
}

// Photo gets the entity contaning this face.
func (f Face) Photo(s *Session) (*Photo, error) {
	return nil, fmt.Errorf("todo")
}
