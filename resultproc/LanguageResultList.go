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

type LanguageResultList struct {
	Logger  zerolog.Logger
	results []LanguageResult
}

func CreateLanguageResultList(results map[string]int32, logger zerolog.Logger) LanguageResultList {
	var resl LanguageResultList
	l := logger.With().Str("function", "CreateLanguageResultList").Logger()

	l.Trace().Msg("Creando logger")
	resl.Logger = logger.With().Str("struct", "ResultList").Logger()

	var min, max int32
	min = math.MaxInt32
	max = 0
	l.Trace().Msg("Calculando minimo and maximo")
	for _, num := range results {
		if num > max {
			max = num
		}
		if num < min {
			min = num
		}
	}
	fmt.Print(min, max, "\n")

	l.Trace().Msg("Parseando mapa a lista")
	for lang, num := range results {
		resl.results = append(resl.results, LanguageResult{
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

func (resl *LanguageResultList) Save(filename string) error {
	l := resl.Logger.With().Str("method", "Save").Logger()

	l.Trace().Msg("Abriendo archivo")
	file, err := os.Create(filename)
	if err != nil {
		l.Error().Err(err).Msg("No se pudo abrir ni crear archivo!")
		return err
	}
	defer file.Close()

	l.Trace().Msg("Guardando resultando")
	for _, res := range resl.results {
		err = res.Save(file)
		if err != nil {
			l.Error().Err(err).Msg("No se pudo guardar resultado!")
			return err
		}
	}
	return nil
}

func (resl *LanguageResultList) Graph(htmlname string) error {
	l := resl.Logger.With().Str("method", "Graph").Logger()

	l.Trace().Msg("Crear nueva gráfica")
	bar := charts.NewBar()

	l.Trace().Msg("Configurar opciones")
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

	l.Trace().Msg("Llenar datos")
	Languagelist := resl.getLanguageList()
	if len(Languagelist) > 10 {
		Languagelist = Languagelist[0:10]
	}
	barlist := resl.getBars()
	if len(barlist) > 10 {
		barlist = barlist[0:10]
	}
	bar.SetXAxis(Languagelist).
		AddSeries("Languages", barlist)

	l.Trace().Str("html-file", htmlname).Msg("Crear archivo html")
	f, err := os.Create(htmlname)
	if err != nil {
		l.Error().Err(err).Msg("No se pudo crear archivo html!")
		return err
	}

	l.Trace().Msg("Guardar en archivo")
	bar.Render(f)
	return nil
}

func (resl *LanguageResultList) ScoreSort() {
	l := resl.Logger.With().Str("method", "ScoreSort").Logger()

	l.Trace().Msg("Ordenar puntajes")
	sort.Sort(sort.Reverse(ScoreSort(resl.results)))
}

func (resl *LanguageResultList) NumSort() {
	l := resl.Logger.With().Str("method", "NumSort").Logger()

	l.Trace().Msg("Ordenar por numero de appariciones en github.com")
	sort.Sort(sort.Reverse(NumSort(resl.results)))
}

func (resl *LanguageResultList) getLanguageList() []string {
	l := resl.Logger.With().Str("method", "getLanguageList").Logger()

	l.Trace().Msg("Obtener array de lenguajes")
	var langs []string
	for _, res := range resl.results {
		langs = append(langs, res.Language)
	}

	l.Trace().Msg("Retornar slice de lenguajes")
	return langs
}
func (resl *LanguageResultList) getValueList() []int32 {
	l := resl.Logger.With().Str("method", "getValueList").Logger()

	l.Trace().Msg("Obtener slice de valores")
	var topicnums []int32
	for _, res := range resl.results {
		topicnums = append(topicnums, res.TopicNum)
	}

	l.Trace().Msg("Retornar slice de valores int32")
	return topicnums
}
func (resl *LanguageResultList) getBars() []opts.BarData {
	l := resl.Logger.With().Str("method", "getBars").Logger()

	l.Trace().Msg("Obtener slice de barras")
	bars := make([]opts.BarData, 0)
	for _, num := range resl.getValueList() {
		bars = append(bars, opts.BarData{Value: num})
	}

	l.Trace().Msg("Retornar slice de barras")
	return bars
}
func (resl *LanguageResultList) String() string {
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
