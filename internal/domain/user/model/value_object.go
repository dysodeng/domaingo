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

type Role uint8

const (
	RoleUser  Role = 0
	RoleAdmin Role = 1
)

func (r Role) IsValid() bool {
	return r == RoleUser || r == RoleAdmin
}
