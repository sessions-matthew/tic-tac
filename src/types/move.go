package types

type Move struct {
	X int `json:"x" binding:"required"`
	Y int `json:"y" binding:"required"`
	T string `json:"t" binding:"required"`
}

type MoveNext struct {
	Username string
}
