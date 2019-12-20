package mwgrs

import (
	"trimmer.io/go-xmp/models/xmp_base"
	"trimmer.io/go-xmp/models/xmp_tpg"
	"trimmer.io/go-xmp/xmp"
)

type RegionStruct struct {
	Area         xmpbase.Area  `xmp:"mwg-rs:Area"`
	Type         RegionType    `xmp:"mwg-rs:Type"`
	Name         string        `xmp:"mwg-rs:Name"`
	Description  string        `xmp:"mwg-rs:Description"`
	FocusUsage   FocusUsage    `xmp:"mwg-rs:FocusUsage"`
	BarCodeValue string        `xmp:"mwg-rs:BarCodeValue"`
	// TODO This is still emitted when there are no child nodes
	Extensions   xmp.Extension `xmp:"mwg-rs:Extensions,empty"`
}

type RegionStructList []RegionStruct

func (x RegionStructList) Typ() xmp.ArrayType {
	return xmp.ArrayTypeUnordered
}

func (x RegionStructList) MarshalXMP(e *xmp.Encoder, node *xmp.Node, m xmp.Model) error {
	return xmp.MarshalArray(e, node, x.Typ(), x)
}

func (x *RegionStructList) UnmarshalXMP(d *xmp.Decoder, node *xmp.Node, m xmp.Model) error {
	return xmp.UnmarshalArray(d, node, x.Typ(), x)
}

type RegionInfo struct {
	AppliedToDimensions xmptpg.Dimensions `xmp:"mwg-rs:AppliedToDimensions"`
	RegionList          RegionStructList  `xmp:"mwg-rs:RegionList"`
}

