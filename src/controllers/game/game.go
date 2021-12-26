package game

import (
	"errors"
	"main/src/stores"
	"main/src/types"
	"sort"
)

func NewGame () types.Game {
	var nGame = types.Game {
		Board: [][]string {
			{" ", " ", " "},
				{" ", " ", " "},
				{" ", " ", " "},
			},
		}
	return nGame
}

func NewPlayer (gameid, player string) {
	g := stores.GameStore[gameid]
	if g.Players == nil || len(g.Players) < 2 {
		// search for player in array
		p := sort.Search(len(g.Players), func(i int) bool {
			return g.Players[i] == player
		})

		if g.Players == nil || p == len(g.Players) {
			// add new player
			g.Players = append(g.Players, player)
		}

		if len(g.Players) == 2 {
			// set who goes first
			g.IsReady = true
			g.Turn = g.Players[0]
		}
	}
	stores.GameStore[gameid] = g
}

var (
	ErrMoveTurn = errors.New("not player's turn")
	ErrMoveTaken = errors.New("space is taken")
	ErrGameNotReady = errors.New("game is not ready")
	ErrGameDone = errors.New("game is over")
)

func checkWin (gameid string) bool {
	board := stores.GameStore[gameid].Board
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

func DoMove (gameid string, x,y int, t string) error {
	g := stores.GameStore[gameid]

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
	if g.Turn == g.Players[0] {
		g.Turn = g.Players[1]
	} else {
		g.Turn = g.Players[0]
	}

	g.IsDone = checkWin(gameid)
	
	if g.IsDone {
		g.Winner = t
	}

	stores.GameStore[gameid] = g
	
	return nil
}

func IsReady (gameid string) bool {
	return stores.GameStore[gameid].IsReady
}

func IsDone (gameid string) bool {
	return stores.GameStore[gameid].IsDone
}
