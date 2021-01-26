package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/awesomepatrol/bb-match-history/discord"
	"github.com/awesomepatrol/bb-match-history/graph"
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
		matches, err := stats.QueryPlayerMatchesShort(name)
		if err != nil {
			// FIXME
			c.String(http.StatusNotFound, err.Error())
			return
		}
		c.JSON(http.StatusOK, matches)
	})
	router.GET("/api/player/:name/details", func(c *gin.Context) {
		name := c.Param("name")
		p, err := stats.QueryPlayerByName(name)
		if err != nil {
			// FIXME
			c.String(http.StatusNotFound, err.Error())
			return
		}
		c.JSON(http.StatusOK, p)
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
		if len(defenders) == 0 || len(deaths) == 0 || len(builders) == 0 {
			c.Status(http.StatusNotFound)
			return
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
	router.GET("/api/graph/player/:name", func(c *gin.Context) {
		name := c.Param("name")
		if name == "" {
			c.String(http.StatusBadRequest, "Missing name parameter")
			return
		}
		matches, err := stats.QueryPlayerMatchesShort(name)
		if err != nil {
			// FIXME
			c.String(http.StatusNotFound, err.Error())
			return
		}
		err = graph.RenderPlayerELO(matches, c.Writer)
		if err != nil {
			// FIXME
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.Status(http.StatusOK)
	})
	router.GET("/api/graph/difficulty/:n", func(c *gin.Context) {
		n, err := strconv.Atoi(c.Param("n"))
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		err = graph.RenderDifficultyBreakdown(c.Writer, time.Now().AddDate(0, 0, -n))
		if err != nil {
			// FIXME
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.Status(http.StatusOK)
	})
	router.GET("/api/stats/ups/avg/:n", func(c *gin.Context) {
		n, err := strconv.Atoi(c.Param("n"))
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		ups, err := stats.GetMatchesAverageUPS(time.Now().AddDate(0, 0, -n))
		if err != nil {
			// FIXME
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.String(http.StatusOK, fmt.Sprintf("%.2f", ups))
	})
	router.GET("/api/graph/ups/:n", func(c *gin.Context) {
		n, err := strconv.Atoi(c.Param("n"))
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		err = graph.RenderHistogramUPS(c.Writer, time.Now().AddDate(0, 0, -n))
		if err != nil {
			// FIXME
			log.Println(err)
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.Status(http.StatusOK)
	})
	router.GET("/api/graph/length/:n", func(c *gin.Context) {
		n, err := strconv.Atoi(c.Param("n"))
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		err = graph.RenderScatterGameLength(c.Writer, time.Now().AddDate(0, 0, -n))
		if err != nil {
			// FIXME
			log.Println(err)
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.Status(http.StatusOK)
	})
	router.Run(addr)
}
