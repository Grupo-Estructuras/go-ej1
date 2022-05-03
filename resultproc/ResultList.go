package resultproc

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/rs/zerolog"
)

type ResultList struct {
	Logger  zerolog.Logger
	results []Result
}

func CreateResultList(results map[string]int32, logger zerolog.Logger) ResultList {
	var resl ResultList
	l := logger.With().Str("function", "CreateResultList").Logger()

	l.Trace().Msg("Creating logger for result list")
	resl.Logger = logger.With().Str("struct", "ResultList").Logger()

	var min, max int32
	min = math.MaxInt32
	max = 0
	l.Trace().Msg("Calculating min and max")
	for _, num := range results {
		if num > max {
			max = num
		}
		if num < min {
			min = num
		}
	}
	fmt.Print(min, max, "\n")

	l.Trace().Msg("Parsing map to result list")
	for lang, num := range results {
		resl.results = append(resl.results, Result{
			Logger:   logger.With().Str("object", "Result").Logger(),
			Min:      min,
			Max:      max,
			TopicNum: num,
			Language: lang,
		})
	}

	l.Trace().Msg("EXIT")
	return resl
}

func (resl *ResultList) Save(filename string) error {
	l := resl.Logger.With().Str("method", "Save").Logger()

	l.Trace().Msg("Opening file")
	file, err := os.Create(filename)
	if err != nil {
		l.Error().Err(err).Msg("Could not create/open file!")
		return err
	}
	defer file.Close()

	l.Trace().Msg("Saving results")
	for _, res := range resl.results {
		err = res.Save(file)
		if err != nil {
			l.Error().Err(err).Msg("Could not save result!")
			return err
		}
	}
	return nil
}

func (resl *ResultList) Graph(htmlname string) error {
	l := resl.Logger.With().Str("method", "Graph").Logger()

	l.Trace().Msg("Create new bar")
	bar := charts.NewBar()

	l.Trace().Msg("Set new bar options")
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Top 20 tiobe en Github",
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:  "slider",
			Start: 0,
			End:   100,
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "1920px",
			Height: "600px",
		}),
	)

	l.Trace().Msg("Fill with data")
	bar.SetXAxis(resl.getLanguageList()[0:10]).
		AddSeries("Lenguajes tiobe", resl.getBars()[0:10])

	l.Trace().Str("html-file", htmlname).Msg("Create html file")
	f, err := os.Create(htmlname)
	if err != nil {
		l.Error().Err(err).Msg("Could not create html file!")
		return err
	}

	l.Trace().Msg("Render to file")
	bar.Render(f)
	return nil
}

func (resl *ResultList) ScoreSort() {
	l := resl.Logger.With().Str("method", "ScoreSort").Logger()

	l.Trace().Msg("Sorting by score")
	sort.Sort(sort.Reverse(ScoreSort(resl.results)))
}

func (resl *ResultList) NumSort() {
	l := resl.Logger.With().Str("method", "NumSort").Logger()

	l.Trace().Msg("Sorting by number of appearance on github.com")
	sort.Sort(sort.Reverse(NumSort(resl.results)))
}

func (resl *ResultList) getLanguageList() []string {
	l := resl.Logger.With().Str("method", "getLanguageList").Logger()

	l.Trace().Msg("Get string slice of languages")
	var langs []string
	for _, res := range resl.results {
		langs = append(langs, res.Language)
	}

	l.Trace().Msg("Return slice of language string")
	return langs
}
func (resl *ResultList) getValueList() []int32 {
	l := resl.Logger.With().Str("method", "getValueList").Logger()

	l.Trace().Msg("Get string slice of values")
	var topicnums []int32
	for _, res := range resl.results {
		topicnums = append(topicnums, res.TopicNum)
	}

	l.Trace().Msg("Return slice of values int32")
	return topicnums
}
func (resl *ResultList) getBars() []opts.BarData {
	l := resl.Logger.With().Str("method", "getBars").Logger()

	l.Trace().Msg("Get slice of bars")
	bars := make([]opts.BarData, 0)
	for _, num := range resl.getValueList() {
		bars = append(bars, opts.BarData{Value: num})
	}

	l.Trace().Msg("Return slice of bars")
	return bars
}
func (resl *ResultList) String() string {
	if resl == nil {
		return ""
	}
	var sb strings.Builder
	for _, res := range resl.results {
		sb.WriteString(res.String())
		sb.WriteString("\n")
	}
	return sb.String()
}
