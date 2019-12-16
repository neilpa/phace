// Package mwgrs implements Image Region Metadata as defined by the
// Metadata Working Group (MWG). The ExifTool docs contain a good
// description of the schema:
//
// https://exiftool.org/TagNames/MWG.html#Regions
package mwgrs

import (
	"fmt"

	"trimmer.io/go-xmp/xmp"
)

var (
	NsMwgRs = xmp.NewNamespace("mwg-rs", "http://www.metadataworkinggroup.com/schemas/regions/", NewModel)
)

func init() {
	xmp.Register(NsMwgRs, xmp.XmpMetadata)
}

func NewModel(name string) xmp.Model {
	return &Regions{}
}

func MakeModel(d *xmp.Document) (*Regions, error) {
	m, err := d.MakeModel(NsMwgRs)
	if err != nil {
		return nil, err
	}
	x, _ := m.(*Regions)
	return x, nil
}

func FindModel(d *xmp.Document) *Regions {
	if m := d.FindModel(NsMwgRs); m != nil {
		return m.(*Regions)
	}
	return nil
}

type Regions struct {
	Regions RegionInfo `xmp:"mwg-rs:Regions"`
}

func (x Regions) Can(nsName string) bool {
	return NsMwgRs.GetName() == nsName
}

func (x Regions) Namespaces() xmp.NamespaceList {
	return xmp.NamespaceList{NsMwgRs}
}

func (x *Regions) SyncModel(d *xmp.Document) error {
	return nil
}

func (x *Regions) SyncFromXMP(d *xmp.Document) error {
	return nil
}

func (x Regions) SyncToXMP(d *xmp.Document) error {
	return nil
}

func (x *Regions) CanTag(tag string) bool {
	_, err := xmp.GetNativeField(x, tag)
	return err == nil
}

func (x *Regions) GetTag(tag string) (string, error) {
	if v, err := xmp.GetNativeField(x, tag); err != nil {
		return "", fmt.Errorf("%s: %v", NsMwgRs.GetName(), err)
	} else {
		return v, nil
	}
}

func (x *Regions) SetTag(tag, value string) error {
	if err := xmp.SetNativeField(x, tag, value); err != nil {
		return fmt.Errorf("%s: %v", NsMwgRs.GetName(), err)
	}
	return nil
}
