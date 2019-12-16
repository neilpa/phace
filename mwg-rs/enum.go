package mwgrs

type RegionType string

const (
	TypeFace    RegionType = "Face"
	TypePet     RegionType = "Pet"
	TypeFocus   RegionType = "Focus"
	TypeBarCode RegionType = "BarCode"
)

type FocusUsage string

const (
	EvaluatedUsed FocusUsage = "EvaluatedUsed"
	EvaluatedNotUsed FocusUsage = "EvaluatedNotUsed"
	NotEvaluatedNotUsed FocusUsage = "NotEvaluatedNotUsed"
)
