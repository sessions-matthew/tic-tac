package routes

import (
	"encoding/json"
	"fmt"
	"main/src/controllers/game"
	"main/src/stores"
	"main/src/types"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nsqio/go-nsq"
)

const nsqd = "localhost:4150"
const nsqlookupd = "localhost:4160"

func Game(c *gin.Context) {
	gameid := c.Param("gameid")
	store := stores.GameStore

	if val, ok := store[gameid]; ok {
		c.JSON(200, val)
	} else {
		nGame := game.NewGame()
		store[gameid] = nGame
		c.JSON(200, nGame)
	}
}

func NewGame(c *gin.Context) {
	username := c.PostForm("username")
	gameid := c.PostForm("gameid")
	fmt.Println(username, gameid)
	
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

func createTopic(topic string) {
	valid := nsq.IsValidTopicName(topic)
	if( valid ) {
		resp,_ := http.Post(
			"http://localhost:4151/topic/create?" + url.Values{"topic": {topic}}.Encode(),
			"application/x-www-form-urlencoded",
			nil,
		)
		fmt.Println(resp)
	}
}

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

func connectNsq(gameid, player string, conn *websocket.Conn) (*nsq.Consumer, *nsq.Producer, error) {
	config := nsq.NewConfig()
	consumer, err1 := nsq.NewConsumer(gameid, player, config)
	if err1 != nil {
		return nil, nil, err1
	}
	
	producer, err2 := nsq.NewProducer(nsqd, config)
	if err2 != nil {
		return nil, nil, err2
	}

	consumer.AddHandler(&myHandler{Socket: conn,})
	err3 := consumer.ConnectToNSQLookupd("localhost:4161")
	if err3 != nil {
		fmt.Println("connect", err3)
		return nil, nil, err3
	}

	return consumer, producer, nil
}

func GameSocket(c *gin.Context) {
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

	// nsq
	createTopic(gameid)
	consumer, producer, err := connectNsq(gameid, player, conn)

	// send player connected event
	s0, _ := json.Marshal(types.PlayerEvent{
		Type: "player",
		Event: "connected",
		Content: player,
	})
	producer.Publish(gameid, s0)

	// create game instance if not exists
	if _, ok := stores.GameStore[gameid]; !ok {
		stores.GameStore[gameid] = game.NewGame()
	}

	// maintain list of players
	game.NewPlayer(gameid, player)

	for {
		var jsonM types.Move
		err := conn.ReadJSON(&jsonM)

		if err != nil {
			fmt.Println("error happened", err)
			break
		}

		err0  := game.DoMove(gameid, jsonM.X, jsonM.Y, jsonM.T)
		if err0 == nil {
			// push move event to clients
			jsonR := types.MoveEvent{
				Type: "move",
				Event: "move",
				Content: jsonM,
			}
			
			s, _ := json.Marshal(jsonR)
			producer.Publish(gameid, s)

			if game.IsDone(gameid) {
				// push game over event to clients
				jsonR := types.GameEvent{
					Type: "game",
					Event: "over",
					Content: jsonM.T,
				}
				s, _ := json.Marshal(jsonR)				
				producer.Publish(gameid, s)
			}

			// log moves
			fmt.Println("gameid:", gameid)
			fmt.Println("move:", string(s))			
		} else {
			fmt.Println(err0)
		}
	}

	// player disconnected
	s, _ := json.Marshal(types.PlayerEvent{
		Type: "player",
		Event: "disconnect",
		Content: player,
	})
	producer.Publish(gameid, s);
	
	consumer.DisconnectFromNSQLookupd("localhost:4161")
	consumer.Stop()
	conn.Close()
}
