package server

import (
	"net/http"

	"github.com/awesomepatrol/bb-match-history/discord"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func OpenHTTP(addr string) {
	router = gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/recent")
	})
	router.GET("/site", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/recent")
	})
	router.GET("/site/:x", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/"+c.Param("x"))
	})
	router.Use(static.Serve("/", static.LocalFile("assets/public", false)))

	api := router.Group("/api")
	api.Use(gzip.Gzip(gzip.DefaultCompression))
	api.GET("/player/:name/history", PlayerHistory)
	api.GET("/player/:name/details", PlayerDetails)
	api.GET("/match/short/:id", MatchShort)
	api.GET("/match/long/:id", MatchLong)
	api.GET("/match/history", MatchHistory)
	api.GET("/match/current", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/match/current/casual")
	})
	api.GET("/match/current/casual", func(c *gin.Context) {
		c.JSON(http.StatusOK, discord.GetCurrentCasual())
	})
	api.GET("/match/current/tournament", func(c *gin.Context) {
		c.JSON(http.StatusOK, discord.GetCurrentTournament())
	})
	api.GET("/match/mvp", MatchMVP)
	api.GET("/elo/:limit", TopELO)
	api.GET("/csv/match/all", MatchCSV)
	api.GET("/graph/player/:name", GraphPlayerELO)
	api.GET("/graph/difficulty/:n", GraphRecentDifficulty)
	api.GET("/stats/ups/avg/:n", RecentUPS)
	api.GET("/graph/ups/:n", GraphRecentUPS)
	api.GET("/graph/length/:n", GraphRecentMatchLength)
	api.GET("/graph/players/:n", GraphRecentPlayerCount)
	api.GET("/graph/evos/:n", GraphLimitEvos)
	router.Run(addr)
}
