package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/awesomepatrol/bb-match-history/graph"
	"github.com/awesomepatrol/bb-match-history/stats"
	"github.com/gin-gonic/gin"
)

func PlayerHistory(c *gin.Context) {
	name := c.Param("name")
	matches, err := stats.QueryPlayerMatchesShort(name)
	if err != nil {
		// FIXME
		c.String(http.StatusNotFound, err.Error())
		return
	}
	c.JSON(http.StatusOK, matches)
}

func PlayerDetails(c *gin.Context) {
	name := c.Param("name")
	p, err := stats.QueryPlayerByName(name)
	if err != nil {
		// FIXME
		c.String(http.StatusNotFound, err.Error())
		return
	}
	c.JSON(http.StatusOK, p)
}

func MatchShort(c *gin.Context) {
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
}

func MatchLong(c *gin.Context) {
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
}

func MatchHistory(c *gin.Context) {
	matches, err := stats.QueryMatchAll()
	if err != nil {
		// FIXME
		c.String(http.StatusNotFound, err.Error())
		return
	}
	c.JSON(http.StatusOK, matches)
}

func MatchMVP(c *gin.Context) {
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
}

func TopELO(c *gin.Context) {
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
}

func MatchCSV(c *gin.Context) {
	c.Status(http.StatusOK)
	err := stats.GetMatchWithFeedsAsCSV(c.Writer)
	if err != nil {
		log.Println("csv dump failed:", err)
	}
}

func GraphPlayerELO(c *gin.Context) {
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
}

func GraphRecentDifficulty(c *gin.Context) {
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
}

func RecentUPS(c *gin.Context) {
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
}

func GraphRecentUPS(c *gin.Context) {
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
}

func GraphRecentMatchLength(c *gin.Context) {
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
}

func GraphRecentPlayerCount(c *gin.Context) {
	n, err := strconv.Atoi(c.Param("n"))
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	err = graph.RenderScatterPlayerCount(c.Writer, time.Now().AddDate(0, 0, -n))
	if err != nil {
		// FIXME
		log.Println(err)
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Status(http.StatusOK)
}

func GraphLimitEvos(c *gin.Context) {
	n, err := strconv.Atoi(c.Param("n"))
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	err = graph.RenderEvoComp(c.Writer, n)
	if err != nil {
		// FIXME
		log.Println(err)
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Status(http.StatusOK)
}
