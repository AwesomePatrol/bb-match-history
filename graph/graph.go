package graph

import (
	"fmt"
	"io"
	"math"
	"time"

	"github.com/awesomepatrol/bb-match-history/stats"
	"github.com/awesomepatrol/bb-match-history/stats/const/difficulty"
	chart "github.com/wcharczuk/go-chart/v2"
)

func RenderHistogramUPS(w io.Writer, after time.Time) error {
	avg, err := stats.GetMatchesAverageUPSAll(after)
	if err != nil {
		return err
	}
	clampMin, clampMax := 53, 61
	buckets := make([]float64, clampMax-clampMin)
	for _, aa := range avg {
		a := int(math.Round(aa))
		if a >= clampMin && a < clampMax {
			buckets[a-clampMin]++
		}
	}
	labels := make([]float64, clampMax-clampMin)
	ticks := make([]chart.Tick, clampMax-clampMin+2)
	ticks[0] = chart.Tick{
		Value: float64(clampMin - 1),
		Label: fmt.Sprint(clampMin - 1),
	}
	for i := clampMin; i < clampMax; i++ {
		labels[i-clampMin] = float64(i)
		ticks[i-clampMin+1] = chart.Tick{
			Value: float64(i),
			Label: fmt.Sprint(i),
		}
	}
	ticks[len(ticks)-1] = chart.Tick{
		Value: float64(clampMax),
		Label: fmt.Sprint(clampMax),
	}
	graph := chart.Chart{
		Height: 360,
		Width:  360,
		XAxis: chart.XAxis{
			Name: "UPS",
			Range: &chart.ContinuousRange{
				Min: float64(clampMin - 1),
				Max: float64(clampMax + 1),
			},
			Ticks: ticks,
		},
		YAxis: chart.YAxis{
			ValueFormatter: chart.IntValueFormatter,
		},
		Series: []chart.Series{
			chart.HistogramSeries{
				InnerSeries: chart.ContinuousSeries{
					XValues: labels,
					YValues: buckets,
				},
			},
		},
	}
	return graph.Render(chart.PNG, w)
}

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
	if len(matches) == 0 {
		return fmt.Errorf("no matches to render")
	}
	x := make([]time.Time, len(matches)+1)
	y := make([]float64, len(matches)+1)
	x[0] = matches[0].Match.Start
	y[0] = float64(matches[0].BeforeELO)
	for i, m := range matches {
		x[i+1] = m.Match.End
		y[i+1] = float64(m.BeforeELO + m.GainELO)
	}
	mainSeries := chart.TimeSeries{
		Style: chart.Style{
			StrokeColor: chart.ColorAlternateGray,
			StrokeWidth: 3.0,
		},
		XValues: x,
		YValues: y,
	}
	minSeries := &chart.MinSeries{
		Style: chart.Style{
			StrokeColor:     chart.ColorAlternateGray,
			StrokeDashArray: []float64{4.0, 4.0},
		},
		InnerSeries: mainSeries,
	}

	maxSeries := &chart.MaxSeries{
		Style: chart.Style{
			StrokeColor:     chart.ColorAlternateGray,
			StrokeDashArray: []float64{4.0, 4.0},
		},
		InnerSeries: mainSeries,
	}
	graph := chart.Chart{
		Height: 300,
		Width:  1000,
		XAxis: chart.XAxis{
			ValueFormatter: func(v interface{}) string {
				if typed, isTyped := v.(time.Time); isTyped {
					return typed.Format("Jan 2")
				}
				if typed, isTyped := v.(float64); isTyped {
					return time.Unix(0, int64(typed)).Format("Jan 2")
				}
				return "---"
			},
			GridMajorStyle: chart.Style{
				StrokeColor: chart.ColorAlternateGray,
				StrokeWidth: 1.0,
			},
			GridMinorStyle: chart.Style{
				StrokeColor: chart.ColorLightGray,
				StrokeWidth: 1.0,
			},
		},
		YAxis: chart.YAxis{
			ValueFormatter: chart.IntValueFormatter,
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
