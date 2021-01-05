package server

import (
	"log"
	"net/http"
	"strconv"

	"github.com/awesomepatrol/bb-match-history/discord"
	"github.com/awesomepatrol/bb-match-history/stats"

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
	router.GET("/api/player/:name/history", func(c *gin.Context) {
		name := c.Param("name")
		matches, err := stats.QueryPlayerMatches(name)
		if err != nil {
			// FIXME
			c.String(http.StatusNotFound, err.Error())
			return
		}
		c.JSON(http.StatusOK, matches)
	})
	router.GET("/api/player/:name/elo", func(c *gin.Context) {
		name := c.Param("name")
		elo, err := stats.QueryPlayerELO(name)
		if err != nil {
			// FIXME
			c.String(http.StatusNotFound, err.Error())
			return
		}
		c.JSON(http.StatusOK, &stats.Player{Name: name, ELO: elo})
	})
	router.GET("/api/match/short/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		match, err := stats.QueryMatchShort(id)
		if err != nil {
			// FIXME
			c.String(http.StatusNotFound, err.Error())
			return
		}
		c.JSON(http.StatusOK, match)
	})
	router.GET("/api/match/long/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		}

		match, err := stats.QueryMatchLong(id)
		if err != nil {
			// FIXME
			c.String(http.StatusNotFound, err.Error())
			return
		}
		c.JSON(http.StatusOK, match)
	})
	router.GET("/api/match/history", func(c *gin.Context) {
		matches, err := stats.QueryMatchAll()
		if err != nil {
			// FIXME
			c.String(http.StatusNotFound, err.Error())
			return
		}
		c.JSON(http.StatusOK, matches)
	})
	router.GET("/api/match/current", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/api/match/current/casual")
	})
	router.GET("/api/match/current/casual", func(c *gin.Context) {
		c.JSON(http.StatusOK, discord.GetCurrentCasual())
	})
	router.GET("/api/match/current/tournament", func(c *gin.Context) {
		c.JSON(http.StatusOK, discord.GetCurrentTournament())
	})
	router.GET("/api/match/mvp", func(c *gin.Context) {
		defenders, err := stats.QueryGlobalMVP("Defender")
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
		}
		deaths, err := stats.QueryGlobalMVP("Deaths")
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
		}
		builders, err := stats.QueryGlobalMVP("Builder")
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
		}
		c.JSON(http.StatusOK, struct {
			Defenders []stats.MVPquery
			Deaths    []stats.MVPquery
			Builders  []stats.MVPquery
		}{
			Defenders: defenders,
			Deaths:    deaths,
			Builders:  builders,
		})
	})
	router.GET("/api/elo/:limit", func(c *gin.Context) {
		limit, err := strconv.Atoi(c.Param("limit"))
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		players, err := stats.QueryTopPlayersByELO(limit)
		if err != nil {
			// FIXME
			c.String(http.StatusNotFound, err.Error())
			return
		}
		c.JSON(http.StatusOK, players)
	})
	router.GET("/api/csv/match/all", func(c *gin.Context) {
		c.Status(http.StatusOK)
		err := stats.GetMatchWithFeedsAsCSV(c.Writer)
		if err != nil {
			log.Println("csv dump failed:", err)
		}
	})
	router.Run(addr)
}
