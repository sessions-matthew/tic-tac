package game

import (
	"errors"
	"fmt"
	"main/src/types"
)

type Player struct {
	Username string `json:"username"`
	EventSource chan interface{} `json:"-"`
	Nth int
}

type Game struct {
	IsDone bool `json:"isDone" binding:"required"`
	IsReady bool `json:"isReady" binding:"required"`
	Winner string `json:"winner" binding:"required"`
	Turn string `json:"turn" binding:"required"`
	Players map[string] Player `json:"players" binding:"required"`
	Board [][]string `json:"board" binding:"required"`
}

type GameStore map[string] Game

func (store GameStore) GameExists(gameid string) bool {
	_, ok := store[gameid]
	return ok
}

func (store GameStore) NewGame (gameid string) {
	var nGame = Game {
		Board: [][]string {
			{" ", " ", " "},
				{" ", " ", " "},
				{" ", " ", " "},
			},
		Players: make(map[string] Player),
	}
	store[gameid] = nGame
}

func (store GameStore) NewPlayer (gameid, player string) {
	g := store[gameid]
	fmt.Println("HERE", len(g.Players))
	if g.Players == nil || len(g.Players) < 2 {
		_, ok := g.Players[player]

		if !ok {
			g.Players[player] = Player { player, make(chan interface{}), len(g.Players)}
		}

		if len(g.Players) == 2 {
			// set who goes first
			g.IsReady = true
			g.Turn = player
		}
	}

	fmt.Println("game>>", g)
	store[gameid] = g
}

var (
	ErrMoveTurn = errors.New("not player's turn")
	ErrMoveTaken = errors.New("space is taken")
	ErrGameNotReady = errors.New("game is not ready")
	ErrGameDone = errors.New("game is over")
)

func (store GameStore) checkWin (gameid string) bool {
	board := store[gameid].Board
	winner := false
	for i := 0; i < 3; i++ {
		if (board[i][0] != " " &&
			board[i][0] == board[i][1] &&
			board[i][0] == board[i][2]) ||
			(board[0][i] != " " &&
				board[0][i] == board[1][i] &&
				board[0][i] == board[2][i]) ||
			(board[1][1] != " " &&
				board[1][1] == board[2-i][2] &&
				board[1][1] == board[i][0]){
			winner =  true
		}
	}
	return winner
}

func (store GameStore) GetNextPlayer (gameid, username string) string {
	g := store[gameid]
	next := (g.Players[g.Turn].Nth + 1) % (len(g.Players))
	fmt.Println(">> next", next)
	for _, p := range g.Players {
		if p.Nth == next {
			return p.Username
		}
	}
	return ""
}

func (store GameStore) DoMove (gameid string, x,y int, t string) error {
	g := store[gameid]

	if !g.IsReady {
		return ErrGameNotReady
	}

	if g.IsDone {
		return ErrGameDone
	}
	
	// not player's turn
	if g.Turn != t {
		return ErrMoveTurn
	}
	
	if g.Board[y][x] != " " {
		return ErrMoveTaken
	}
	
	// place player's piece
	g.Board[y][x] = t;
	
	// cycle between players
	g.Turn = store.GetNextPlayer(gameid, g.Turn)
	
	g.IsDone = store.checkWin(gameid)
	
	if g.IsDone {
		g.Winner = t
	}

	store[gameid] = g
	
	return nil
}

func (store GameStore) IsReady (gameid string) bool {
	return store[gameid].IsReady
}

func (store GameStore) IsDone (gameid string) bool {
	return store[gameid].IsDone
}

func (store GameStore) HasGame (gameid string) bool {
	_, ok := store[gameid]
	return ok
}

func (store GameStore) GameProcess (eventsFromClient chan interface{}) {
	for {
		event := <- eventsFromClient
		fmt.Println(">> event pre type: ", event)

		switch event.(type) {
		case types.GameConnect:
			e := event.(types.GameConnect)
			fmt.Println("player connected:", e)
			for _, p := range store[e.GameId].Players {
				p.EventSource <- e
			}
			break
			
		case types.GameDisconnect:
			e := event.(types.GameDisconnect)
			fmt.Println("player disconnected:", e)
			break

		case *types.GameRequest:
			fmt.Println(">> requesting game instance")
			e := event.(*types.GameRequest)

			if !store.GameExists(e.GameId) {
				fmt.Println(">> creating new game instance")
				store.NewGame(e.GameId)
			}

			store.NewPlayer(e.GameId, e.Username)

			e.Complete <- true
			break

		case types.GameMove:
			e := event.(types.GameMove)

			err0  := store.DoMove(e.GameId, e.Move.X, e.Move.Y, e.Move.T)
			if err0 == nil {
				// push move event to clients
				for _, player := range store[e.GameId].Players {
					player.EventSource <- e.Move
				}
				if store.IsDone(e.GameId) {
					// push game over event to clients
					msg := types.GameOver {
						Winner: store[e.GameId].Winner,
						}
					for _, player := range store[e.GameId].Players {
						player.EventSource <- msg
					}
				} else {
					// game is not done, send who is next
					for _, player := range store[e.GameId].Players {
						player.EventSource <- types.MoveNext{store[e.GameId].Turn}
					}
				}

				// log moves
				fmt.Println("gameid:", e.GameId)
				fmt.Println("move:", e.Move)
			} else {
				fmt.Println(err0)
			}
			
			break
		}
	}
}
