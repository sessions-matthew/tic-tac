package types

type PlayerEvent struct {
	Type string `json:"type" binding:"required"`
	Event string `json:"event" binding:"required"`
	Content string `json:"content" binding:"required"`
}
