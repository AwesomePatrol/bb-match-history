package graph

import (
	"io"
	"time"

	"github.com/awesomepatrol/bb-match-history/stats"
	chart "github.com/wcharczuk/go-chart/v2"
)

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
		XAxis: chart.XAxis{
			TickPosition: chart.TickPositionBetweenTicks,
		},
		YAxis: chart.YAxis{
			GridMajorStyle: chart.Style{
				StrokeColor: chart.ColorAlternateGray,
				StrokeWidth: 1.0,
			},
			GridLines: []chart.GridLine{
				{Value: 800},
			},
		},
		Series: []chart.Series{
			mainSeries,
			minSeries,
			maxSeries,
			chart.LastValueAnnotationSeries(minSeries),
			chart.LastValueAnnotationSeries(maxSeries),
		},
	}
	return graph.Render(chart.PNG, w)
}
