package main

import (
	"fmt"
	"sort"
	"net/http"
	"net/url"
	"encoding/json"
)

import "github.com/gin-gonic/gin"
import "github.com/gin-contrib/static"
import "github.com/gorilla/websocket"
import "github.com/gin-contrib/cors"
import "github.com/nsqio/go-nsq"

var wsu = websocket.Upgrader {
	ReadBufferSize: 1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return r.Host == "localhost:3000" ||
				r.Host == "localhost:8080"
		},
}

type Move struct {
	X int `json:"x" binding:"required"`
	Y int `json:"y" binding:"required"`
	T string `json:"t" binding:"required"`
}

type Game struct {
	IsDone bool `json:"isDone" binding:"required"`
	IsReady bool `json:"isReady" binding:"required"`
	Winner string `json:"winner" binding:"required"`
	Turn string `json:"turn" binding:"required"`
	Players []string `json:"players" binding:"required"`
	Board [][]string `json:"board" binding:"required"`
}

type PlayerEvent struct {
	Type string `"json:"type" binding:"required"`
	Event string `json:"event" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type GameEvent struct {
	Type string `json:"type" binding:"required" default:"game"`
	Event string `json:"event" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type MoveEvent struct {
	Type string `json:"type" binding:"required" default:"move"`
	Event string `json:"event" binding:"required"`
	Content Move `json:"content" binding:"required"`
}

var GameStore = map[string] Game {}

type myHandler struct {
	Socket *websocket.Conn
}

func (h *myHandler) HandleMessage (m *nsq.Message) error {
	if len(m.Body) == 0 {
		return nil
	}
	// forward message to connected client
	h.Socket.WriteMessage(1, m.Body)
	return nil
}

func checkWin (id string) bool {
	board := GameStore[id].Board
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

func doMove (game *Game, x,y int, t string) string {
	// not player's turn
	if game.Turn != t {
		return "turn"
	}
	// cycle between players
	if game.Turn == game.Players[0] {
		game.Turn = game.Players[1]
	} else {
		game.Turn = game.Players[0]
	}
	// place player's piece
	if game.Board[y][x] == " " {
		game.Board[y][x] = t;
		return "ok"		
	}
	return "taken"
}

func newGame () Game {
	var nGame = Game {
		Board: [][]string {
			{" ", " ", " "},
				{" ", " ", " "},
				{" ", " ", " "},
			},
		}
	return nGame
}

func main() {
	r := gin.Default()
	r.Use(cors.Default())
	r.Use(static.Serve("/", static.LocalFile("./public", false)))
	r.LoadHTMLGlob("templates/*.tmpl")
	
	r.POST("/newgame", func(c *gin.Context) {
		username := c.PostForm("username")
		gameid := c.PostForm("gameid")
		fmt.Println(username, gameid)
		
		c.HTML(http.StatusOK, "game.tmpl", gin.H{
			"username": username,
			"gameid": gameid,
			"player": username,
		})
	})
	
	r.GET("/game/:gameid", func(c *gin.Context) {
		gameid := c.Param("gameid")

		if val, ok := GameStore[gameid]; ok {
			c.JSON(200, val)
		} else {
			nGame := newGame()
			GameStore[gameid] = nGame
			c.JSON(200, nGame)
		}
	})

	r.GET("/listen/:gameid/:player", func(c *gin.Context) {
		wsu.CheckOrigin = func(r *http.Request) bool { return true }
		conn, err := wsu.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println("could not upgrade socket", err)
			return
		}
		
		gameid := c.Param("gameid")
		player := c.Param("player")
		fmt.Println("player connected:", player)

		valid := nsq.IsValidTopicName(gameid)

		if( valid ) {
			resp,_ := http.Post(
				"http://localhost:4151/topic/create?" + url.Values{"topic": {gameid}}.Encode(),
				"application/x-www-form-urlencoded",
				nil,
			)
			fmt.Println(resp)
		}

		config := nsq.NewConfig()

		consumer, err1 := nsq.NewConsumer(gameid, player, config)
		if err1 != nil {
			fmt.Println("consumer", err1)
			return
		}

		producer, err2 := nsq.NewProducer("localhost:4150", config)
		if err2 != nil {
			fmt.Println("producer", err2)
			return
		}

		consumer.AddHandler(&myHandler{Socket: conn,})
		
		err3 := consumer.ConnectToNSQLookupd("localhost:4161")
		if err3 != nil {
			fmt.Println("connect", err3)
			return
		}

		s0, _ := json.Marshal(PlayerEvent{
			Type: "player",
			Event: "connected",
			Content: player,
		})
		producer.Publish(gameid, s0)

		if _, ok := GameStore[gameid]; !ok {
			GameStore[gameid] = newGame()
		}
		
		g := GameStore[gameid]
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
			// commit new game state			
			GameStore[gameid] = g
		}

		for {
			var jsonM Move	
			err := conn.ReadJSON(&jsonM)

			if err != nil {
				fmt.Println("error happened", err)
				break
			}

			game := GameStore[gameid]
			if game.IsReady {
				var status string
				if !GameStore[gameid].IsDone {
					status = doMove(&game, jsonM.X, jsonM.Y, jsonM.T)
				} else {
					status = "done"
				}

				fmt.Println("STATUS", status)

				if status != "turn" {
					game.IsDone = checkWin(gameid)
					GameStore[gameid] = game

					jsonR := MoveEvent{
						Type: "move",
						Event: "move",
						Content: jsonM,
					}
					s, _ := json.Marshal(jsonR)
					producer.Publish(gameid, s)

					if status == "ok" && GameStore[gameid].IsDone {
						jsonR := GameEvent{
							Type: "game",
							Event: "over",
							Content: jsonM.T,
						}
						s, _ := json.Marshal(jsonR)				
						producer.Publish(gameid, s)
					}
					fmt.Println("gameid:", gameid)
					fmt.Println("string:", string(s))
				}
			}
		}

		s, _ := json.Marshal(PlayerEvent{
			Type: "player",
			Event: "disconnect",
			Content: player,
		})
		producer.Publish(gameid, s);
		
		consumer.DisconnectFromNSQLookupd("localhost:4161")
		consumer.Stop()
		conn.Close()
	})

	r.Run("localhost:8080")
}
