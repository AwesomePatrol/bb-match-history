package server

import (
	"net/http"
	"strconv"

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
	router.GET("/api/match/short/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
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
	router.Run(addr)
}
