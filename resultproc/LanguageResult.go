package resultproc

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

type LanguageResult struct {
	Logger   zerolog.Logger
	Min, Max int32
	Language string
	TopicNum int32
	Score    float32
}

type ScoreSort []LanguageResult

func (a ScoreSort) Len() int           { return len(a) }
func (a ScoreSort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ScoreSort) Less(i, j int) bool { return a[i].Score < a[j].Score }

type NumSort []LanguageResult

func (a NumSort) Len() int           { return len(a) }
func (a NumSort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a NumSort) Less(i, j int) bool { return a[i].TopicNum < a[j].TopicNum }

func (res *LanguageResult) Save(file *os.File) error {
	l := res.Logger.With().Str("method", "Save").Str("lang", res.Language).Logger()

	l.Trace().Msg("Intentando guardar resultado")
	_, err := file.WriteString(fmt.Sprintf("%v,%v\n", res.Language, res.TopicNum))
	if err != nil {
		l.Error().Err(err).Msg("No se pudo escribir en archivo")
		return err
	}
	l.Trace().Msg("EXIT")
	return nil
}

func (res *LanguageResult) GetScore() float32 {
	l := res.Logger.With().Str("method", "GetScore").Logger()

	if res.Score == 0 {
		l.Trace().Msg("Calculando puntaje")
		res.Score = float32(res.TopicNum-res.Min) / float32(res.Max-res.Min) * 100
	}
	l.Trace().Msg("Retornando puntaje")
	return res.Score
}

func (res *LanguageResult) String() string {
	if res == nil {
		return ""
	}
	return fmt.Sprintf("%40s, %20f, %20d", res.Language, res.GetScore(), res.TopicNum)
}
