package resultproc

import (
	"os"
	"sort"
	"strings"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/rs/zerolog"
)

type TagResultList struct {
	Logger  zerolog.Logger
	results []TagResult
}

func CreateTagResultList(results map[string]int, logger zerolog.Logger) TagResultList {
	var resl TagResultList
	l := logger.With().Str("function", "CreateTagResultList").Logger()

	l.Trace().Msg("Creating logger for result list")
	resl.Logger = logger.With().Str("struct", "TagResultList").Logger()

	l.Trace().Msg("Parsing map to result list")
	for tag, num := range results {
		resl.results = append(resl.results, TagResult{
			Logger: logger.With().Str("object", "Result").Logger(),
			Tag:    tag,
			Num:    num,
		})
	}

	l.Trace().Msg("EXIT")
	return resl
}

func (resl *TagResultList) Graph(htmlname string) error {
	l := resl.Logger.With().Str("method", "Graph").Logger()

	l.Trace().Msg("Create new bar")
	bar := charts.NewBar()

	l.Trace().Msg("Set new bar options")
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Top 20 tags",
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
	taglist := resl.getTagList()
	if len(taglist) > 10 {
		taglist = taglist[0:10]
	}
	barlist := resl.getBars()
	if len(barlist) > 10 {
		barlist = barlist[0:10]
	}
	bar.SetXAxis(taglist).
		AddSeries("Tags", barlist)

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

func (resl *TagResultList) TagSort() {
	l := resl.Logger.With().Str("method", "TagSort").Logger()

	l.Trace().Msg("Sorting tags")
	sort.Sort(sort.Reverse(TagSort(resl.results)))
}

func (resl *TagResultList) getTagList() []string {
	l := resl.Logger.With().Str("method", "getTagList").Logger()

	l.Trace().Msg("Get string slice of tags")
	var tags []string
	for _, res := range resl.results {
		tags = append(tags, res.Tag)
	}

	l.Trace().Msg("Return slice of tags string")
	return tags
}
func (resl *TagResultList) getValueList() []int {
	l := resl.Logger.With().Str("method", "getValueList").Logger()

	l.Trace().Msg("Get string slice of values")
	var tagnums []int
	for _, res := range resl.results {
		tagnums = append(tagnums, res.Num)
	}

	l.Trace().Msg("Return slice of values int32")
	return tagnums
}
func (resl *TagResultList) getBars() []opts.BarData {
	l := resl.Logger.With().Str("method", "getBars").Logger()

	l.Trace().Msg("Get slice of bars")
	bars := make([]opts.BarData, 0)
	for _, num := range resl.getValueList() {
		bars = append(bars, opts.BarData{Value: num})
	}

	l.Trace().Msg("Return slice of bars")
	return bars
}

func (resl *TagResultList) String() string {
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
