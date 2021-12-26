package types

type Move struct {
	X int `json:"x" binding:"required"`
	Y int `json:"y" binding:"required"`
	T string `json:"t" binding:"required"`
}

type MoveEvent struct {
	Type string `json:"type" binding:"required"`
	Event string `json:"event" binding:"required"`
	Content Move `json:"content" binding:"required"`
}
