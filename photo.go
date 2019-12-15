package main

// Photo record in the library, references an image on disk.
type Photo struct {
	UUID        string `db:"uuid"`
	MasterUUID  string `db:"masterUuid"`
	Fingerprint string `db:"fingerprint"`
	// Path to the image on disk, relative to the library root.
	Path string `db:"imagePath"`

	Height int `db:"masterHeight"`
	Width  int `db:"masterWidth"`

	Orientation    int  `db:"orientation"`
	Type           int  `db:"type"`
	HasAdjustments bool `db:"hasAdjustments"`
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
