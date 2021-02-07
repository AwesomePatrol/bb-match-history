package server

import (
	"net/http"

	"github.com/awesomepatrol/bb-match-history/discord"

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
	router.GET("/api/player/:name/history", PlayerHistory)
	router.GET("/api/player/:name/details", PlayerDetails)
	router.GET("/api/match/short/:id", MatchShort)
	router.GET("/api/match/long/:id", MatchLong)
	router.GET("/api/match/history", MatchHistory)
	router.GET("/api/match/current", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/api/match/current/casual")
	})
	router.GET("/api/match/current/casual", func(c *gin.Context) {
		c.JSON(http.StatusOK, discord.GetCurrentCasual())
	})
	router.GET("/api/match/current/tournament", func(c *gin.Context) {
		c.JSON(http.StatusOK, discord.GetCurrentTournament())
	})
	router.GET("/api/match/mvp", MatchMVP)
	router.GET("/api/elo/:limit", TopELO)
	router.GET("/api/csv/match/all", MatchCSV)
	router.GET("/api/graph/player/:name", GraphPlayerELO)
	router.GET("/api/graph/difficulty/:n", GraphRecentDifficulty)
	router.GET("/api/stats/ups/avg/:n", RecentUPS)
	router.GET("/api/graph/ups/:n", GraphRecentUPS)
	router.GET("/api/graph/length/:n", GraphRecentMatchLength)
	router.GET("/api/graph/players/:n", GraphRecentPlayerCount)
	router.GET("/api/graph/evos/:n", GraphLimitEvos)
	router.Run(addr)
}
