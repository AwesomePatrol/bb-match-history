package graph

import (
	"io"
	"time"

	"github.com/awesomepatrol/bb-match-history/stats"
	"github.com/awesomepatrol/bb-match-history/stats/const/difficulty"
	chart "github.com/wcharczuk/go-chart/v2"
)

func RenderDifficultyBreakdown(w io.Writer, after time.Time) error {
	c := make([]int64, 8)
	for i := 0; i < 8; i++ {
		d := difficulty.Difficulty(i)
		v, err := stats.CountMatchesByDifficulty(d, after)
		if err != nil {
			return err
		}
		c[i] = v
	}
	pie := chart.PieChart{
		Width:  360,
		Height: 360,
		Values: []chart.Value{
			{Value: float64(c[0] + c[1] + c[2]), Label: "Below Normal"},
			{Value: float64(c[3]), Label: "Normal"},
			{Value: float64(c[4] + c[5] + c[6] + c[7]), Label: "Above Normal"},
		},
	}
	return pie.Render(chart.PNG, w)
}

func RenderPlayerELO(matches []*stats.GamePlayer, w io.Writer) error {
	x := make([]time.Time, len(matches))
	y := make([]float64, len(matches))
	for i, m := range matches {
		x[i] = m.Match.End
		y[i] = float64(m.BeforeELO + m.GainELO)
	}
	mainSeries := chart.TimeSeries{
		XValues: x,
		YValues: y,
	}
	minSeries := &chart.MinSeries{
		Style: chart.Style{
			StrokeColor:     chart.ColorAlternateGray,
			StrokeDashArray: []float64{5.0, 5.0},
		},
		InnerSeries: mainSeries,
	}

	maxSeries := &chart.MaxSeries{
		Style: chart.Style{
			StrokeColor:     chart.ColorAlternateGray,
			StrokeDashArray: []float64{5.0, 5.0},
		},
		InnerSeries: mainSeries,
	}
	graph := chart.Chart{
		Height: 120,
		Width:  480,
		XAxis: chart.XAxis{
			TickPosition: chart.TickPositionBetweenTicks,
		},
		Series: []chart.Series{
			mainSeries,
			minSeries,
			maxSeries,
			//chart.LastValueAnnotationSeries(minSeries),
			//chart.LastValueAnnotationSeries(maxSeries),
		},
	}
	return graph.Render(chart.PNG, w)
}
