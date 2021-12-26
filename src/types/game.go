package types

type Game struct {
	IsDone bool `json:"isDone" binding:"required"`
	IsReady bool `json:"isReady" binding:"required"`
	Winner string `json:"winner" binding:"required"`
	Turn string `json:"turn" binding:"required"`
	Players []string `json:"players" binding:"required"`
	Board [][]string `json:"board" binding:"required"`
}

type GameEvent struct {
	Type string `json:"type" binding:"required"`
	Event string `json:"event" binding:"required"`
	Content string `json:"content" binding:"required"`
}
