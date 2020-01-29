package server

import (
	"net/http"
	"strconv"

	"github.com/awesomepatrol/bb-match-history/stats"
	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func OpenHTTP() {
	router = gin.Default()

	router.Static("/assets", "./assets")
	router.GET("/api/player/history/:name", func(c *gin.Context) {
		//name := c.Param("name")
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
	router.GET("/api/match/history/", func(c *gin.Context) {
		matches, err := stats.QueryMatchAll()
		if err != nil {
			// FIXME
			c.String(http.StatusNotFound, err.Error())
			return
		}
		c.JSON(http.StatusOK, matches)
	})
	router.Run(":8080")
}
