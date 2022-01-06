package types

type GameMove struct {
	GameId string `json:"gameId" binding:"required"`
	Move Move
}

type GameOver struct {
	Winner string
}

type GameRequest struct {
	GameId string
	Username string
	Complete chan interface{}
}

type GameConnect struct {
	GameId string
	Username string
}

type GameDisconnect struct {
	GameId string
	Username string
}
