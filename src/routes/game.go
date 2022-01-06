package routes

import (
	"fmt"

	"main/src/types"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const nsqd = "localhost:4150"
const nsqlookupd = "localhost:4160"

func LoadGame(c *gin.Context, events chan interface {}) {
	username := c.PostForm("username")
	gameid := c.PostForm("gameid")

	c.HTML(http.StatusOK, "game.tmpl", gin.H{
		"username": username,
		"gameid": gameid,
		"player": username,
	})
}

var wsu = websocket.Upgrader {
	ReadBufferSize: 1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return r.Host == "localhost:3000" ||
				r.Host == "localhost:8080"
		},
	}

func GameSocket(c *gin.Context, events chan interface {}, gameEvents chan interface {}) {
	// websocket
	wsu.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := wsu.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("could not upgrade socket", err)
		return
	}

	// parameters
	gameid := c.Param("gameid")
	player := c.Param("player")
	
	fmt.Println("player connected:", player)
	fmt.Println("gameid:", gameid)

	events <- types.GameConnect {
		GameId: gameid,
		Username: player,
	}

	incomingJSON := make(chan interface{})

	// get json from websocket and forward messages to channel
	go func() {
		for {
			var jsonM types.Move		
			err := conn.ReadJSON(&jsonM)			
			if err != nil {
				fmt.Println("error happened", err)
				incomingJSON <- err
			} else {
				incomingJSON <- jsonM
			}
		}
  }()

	var connected = true
	for connected {
		select {
		// send anything to the central controller
		// from the socket
		case j := <- incomingJSON:
			switch j.(type) {
			case types.Move:
				m := j.(types.Move)
				events <- types.GameMove {
					GameId: gameid,
					Move: m,
				}
				break
			case error:
				connected = false
				return
			}
			break
		// handle events sent to this player
		// to the socket
		case c := <- gameEvents:
			fmt.Println("controller has sent", c)
			switch c.(type) {
			case types.Move:
				m := c.(types.Move)
				conn.WriteJSON(gin.H {
					"type" : "move",
						"x" : m.X,
						"y" : m.Y,
						"t" : m.T,
					})
				break
				
			case types.MoveNext:
				m := c.(types.MoveNext)
				conn.WriteJSON(gin.H {
					"type" : "movenext",
					"username" : m.Username,
				})				
				break
				
			case types.GameOver:
				m := c.(types.GameOver)
				conn.WriteJSON(gin.H {
					"type" : "game",
					"event" : "over",
					"winner" : m.Winner,
				})
				break
				
			case types.GameConnect:
				m := c.(types.GameConnect)
				conn.WriteJSON(gin.H {
					"type" : "player",
					"event" : "connected",
					"content" : m.Username,
				})				
				break
			}
			break
		}
	}
	fmt.Println("disconnected")

	events <- types.GameDisconnect {
		GameId: gameid,
		Username: player,
	}
	
	conn.Close()
}
