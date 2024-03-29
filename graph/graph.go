package graph

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"time"

	"github.com/awesomepatrol/bb-match-history/stats"
	"github.com/awesomepatrol/bb-match-history/stats/const/difficulty"
	chart "github.com/wcharczuk/go-chart/v2"
)

func shortMonthDay(v interface{}) string {
	if typed, isTyped := v.(time.Time); isTyped {
		return typed.Format("Jan 2")
	}
	if typed, isTyped := v.(float64); isTyped {
		return time.Unix(0, int64(typed)).Format("Jan 2")
	}
	return "---"
}

func matchesGameLengthSeries(matches []stats.Match) []chart.TimeSeries {
	series := make([]chart.TimeSeries, 8)
	for i := range series {
		series[i].Style = chart.Style{
			StrokeWidth: chart.Disabled,
			StrokeColor: chart.Viridis(float64(i), 0, 7),
			DotWidth:    5,
			DotColor:    chart.Viridis(float64(i), 0, 7),
		}
		series[i].Name = difficulty.DifficultyToString(difficulty.Difficulty(i))
	}

	for _, m := range matches {
		series[m.Difficulty].XValues = append(series[m.Difficulty].XValues, m.End)
		series[m.Difficulty].YValues = append(series[m.Difficulty].YValues, m.Length.Minutes())
	}

	return series
}

func matchesPlayerCountSeries(matches []stats.Match, counts []stats.MatchPlayerCount) ([]chart.TimeSeries, error) {
	series := make([]chart.TimeSeries, 8)
	for i := range series {
		series[i].Style = chart.Style{
			StrokeWidth: chart.Disabled,
			StrokeColor: chart.Viridis(float64(i), 0, 7),
			DotWidth:    5,
			DotColor:    chart.Viridis(float64(i), 0, 7),
		}
		series[i].Name = difficulty.DifficultyToString(difficulty.Difficulty(i))
	}

	for i, m := range matches {
		if cntID, mID := counts[i].MatchID, m.ID; cntID != mID {
			return nil, fmt.Errorf("id mismatch, count(%d) vs match(%d)", cntID, mID)
		}
		n := counts[i].PlayerCount
		series[m.Difficulty].XValues = append(series[m.Difficulty].XValues, m.End)
		series[m.Difficulty].YValues = append(series[m.Difficulty].YValues, float64(n))
	}

	return series, nil
}

func seriesScatterGraph(series []chart.TimeSeries, last time.Time) *chart.Chart {
	graphSeries := make([]chart.Series, 9)
	annotations := make([]chart.Value2, 8)
	for i := range series {
		graphSeries[i] = series[i]
		annotations[i] = chart.Value2{
			Style: chart.Style{
				FontSize:  8,
				FillColor: chart.Viridis(float64(i), 0, 7),
			},
			XValue: chart.TimeToFloat64(last),
			YValue: chart.ValueSequence(series[i].YValues...).Average(),
			Label:  "-",
		}
	}
	graphSeries[8] = chart.AnnotationSeries{
		Annotations: annotations,
	}

	graph := chart.Chart{
		Height: 600,
		Width:  1000,
		XAxis: chart.XAxis{
			ValueFormatter: shortMonthDay,
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
		Series: graphSeries,
	}
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}
	return &graph
}

func RenderScatterGameLength(w io.Writer, after time.Time) error {
	matches, err := stats.GetAllMatchesAfter(after)
	if err != nil {
		return err
	}

	series := matchesGameLengthSeries(matches)
	graph := seriesScatterGraph(series, matches[0].End)

	return graph.Render(chart.PNG, w)
}

func RenderEvoComp(w io.Writer, n int) error {
	evos, err := stats.GetAllEvo(n)
	if err != nil {
		return err
	}
	n = len(evos)

	// 0 - Winner
	// 1 - Loser
	series := make([]chart.ContinuousSeries, 2)
	for i := range series {
		series[i].XValues = make([]float64, n)
		series[i].YValues = make([]float64, n)
		series[i].Style = chart.Style{
			StrokeWidth: chart.Disabled,
			DotWidth:    5,
		}
	}
	series[0].Name = "Winner"
	series[0].Style.StrokeColor = chart.ColorRed
	series[0].Style.DotColor = chart.ColorRed
	series[1].Name = "Loser"
	series[1].Style.StrokeColor = chart.ColorBlue
	series[1].Style.DotColor = chart.ColorBlue

	ticks := make([]chart.Tick, n)

	for i, e := range evos {
		w, l := e.North, e.South
		if e.Winner == stats.South {
			w, l = l, w
		}
		i := n - i - 1
		series[0].YValues[i] = float64(w)
		series[0].XValues[i] = float64(i)
		series[1].YValues[i] = float64(l)
		series[1].XValues[i] = float64(i)

		ticks[i].Value = float64(i)
		switch i {
		case 0:
			ticks[i].Label = "oldest"
		case n - 1:
			ticks[i].Label = "newest"
		default:
			ticks[i].Label = strconv.Itoa(int(e.ID))
		}
	}
	graph := chart.Chart{
		Height: 600,
		Width:  1000,
		XAxis: chart.XAxis{
			GridMajorStyle: chart.Style{
				StrokeColor: chart.ColorAlternateGray,
				StrokeWidth: 1.0,
			},
			GridMinorStyle: chart.Style{
				StrokeColor: chart.ColorAlternateGray,
				StrokeWidth: 1.0,
			},
			Ticks: ticks,
		},
		YAxis: chart.YAxis{
			ValueFormatter: chart.IntValueFormatter,
			GridMinorStyle: chart.Style{
				StrokeColor: chart.ColorLightGray,
				StrokeWidth: 1.0,
			},
			Range: &chart.ContinuousRange{
				Min: 0,
				Max: 160,
			},
		},
		Series: []chart.Series{
			series[1],
			series[0],
		},
	}
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}

	return graph.Render(chart.PNG, w)
}

func RenderScatterPlayerCount(w io.Writer, after time.Time) error {
	matches, err := stats.GetAllMatchesAfter(after)
	if err != nil {
		return err
	}

	cnt, err := stats.GetAllMatchesPlayerCount(after)
	if err != nil {
		return err
	}
	if cntLen, matchesLen := len(cnt), len(matches); cntLen != matchesLen {
		return fmt.Errorf("length mismatch between count (%d) and matches (%d)", cntLen, matchesLen)
	}

	series, err := matchesPlayerCountSeries(matches, cnt)
	if err != nil {
		return err
	}
	graph := seriesScatterGraph(series, matches[0].End)

	return graph.Render(chart.PNG, w)
}

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
	x := make([]time.Time, 2*len(matches))
	y := make([]float64, 2*len(matches))
	elo := matches[len(matches)-1].BeforeELO
	for i := range matches { // reverse order
		m := matches[len(matches)-1-i]
		x[i*2] = m.Match.Start
		y[i*2] = float64(elo)
		elo += m.GainELO
		x[i*2+1] = m.Match.End
		y[i*2+1] = float64(elo)
	}
	last_match := matches[0]
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
			ValueFormatter: shortMonthDay,
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
			chart.AnnotationSeries{
				Annotations: []chart.Value2{
					{
						XValue: float64(last_match.Match.End.UnixNano()),
						YValue: float64(last_match.BeforeELO + last_match.GainELO),
						Label:  fmt.Sprint(last_match.BeforeELO + last_match.GainELO),
					},
				},
			},
			//chart.LastValueAnnotationSeries(minSeries),
			//chart.LastValueAnnotationSeries(maxSeries),
		},
	}
	return graph.Render(chart.PNG, w)
}
