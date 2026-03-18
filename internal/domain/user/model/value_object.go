package model

type Gender uint8

const (
	Unspecified Gender = iota
	Male
	Female
)

func (g Gender) String() string {
	switch g {
	case Male:
		return "Male"
	case Female:
		return "Female"
	default:
		return "Unspecified"
	}
}

func (g Gender) IsValid() bool {
	return g == Male || g == Female || g == Unspecified
}

func (g Gender) ToInt() int {
	return int(g)
}
