package main

import (
	"main/src/routes"
)

import "github.com/gin-gonic/gin"
import "github.com/gin-contrib/static"
import "github.com/gin-contrib/cors"

func main() {
	r := gin.Default()
	r.Use(cors.Default())
	r.Use(static.Serve("/", static.LocalFile("../public", false)))
	r.LoadHTMLGlob("../templates/*.tmpl")
	
	r.POST("/game", routes.NewGame)
	r.GET("/game/:gameid", routes.Game)
	r.GET("/game/:gameid/:player", routes.GameSocket)
	
	r.Run("localhost:8080")
}
