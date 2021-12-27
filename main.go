package main

import (
	"main/src/controllers/game"
	"main/src/routes"
	"time"

	"main/src/types"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.Use(cors.Default())
	r.Use(static.Serve("/", static.LocalFile("./public", false)))
	r.LoadHTMLGlob("./templates/*.tmpl")

	eventsFromClient := make(chan interface{})
	gameStore := make(game.GameStore)

	go gameStore.GameProcess(eventsFromClient)
	
	r.POST("/game", func (c *gin.Context) {
		routes.LoadGame(c, eventsFromClient)
	})
	
	r.GET("/game/:gameid", func (c *gin.Context) {
		gameid := c.Param("gameid")

		if gameStore.HasGame(gameid) {
			c.JSON(200, gameStore[gameid])
		} else {
			c.JSON(400, "{}")
		}
	})
	
	r.GET("/game/:gameid/:player", func (c *gin.Context) {
		gameid := c.Param("gameid")
		player := c.Param("player")

		r := types.GameRequest {
			GameId: gameid,
			Username: player,
			Complete: false,
		}
		eventsFromClient <- &r

		for !r.Complete {
			time.Sleep(time.Second/4)
		}

		gameEvents := gameStore[gameid].Players[player].EventSource
		routes.GameSocket(c, eventsFromClient, gameEvents)
	})
	
	r.Run("localhost:8080")
}
