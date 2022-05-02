package resultproc

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

type Result struct {
	Logger   zerolog.Logger
	Min, Max int32
	Language string
	TopicNum int32
	Score    float32
}

type ScoreSort []Result

func (a ScoreSort) Len() int           { return len(a) }
func (a ScoreSort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ScoreSort) Less(i, j int) bool { return a[i].Score < a[j].Score }

type NumSort []Result

func (a NumSort) Len() int           { return len(a) }
func (a NumSort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a NumSort) Less(i, j int) bool { return a[i].TopicNum < a[j].TopicNum }

func (res *Result) Save(file *os.File) error {
	l := res.Logger.With().Str("method", "Save").Str("lang", res.Language).Logger()

	l.Trace().Msg("Trying to save result")
	_, err := file.WriteString(fmt.Sprintf("%v,%v\n", res.Language, res.TopicNum))
	if err != nil {
		l.Error().Err(err).Msg("Could not write to file")
		return err
	}
	l.Trace().Msg("EXIT")
	return nil
}

func (res *Result) GetScore() float32 {
	l := res.Logger.With().Str("method", "GetScore").Logger()

	if res.Score == 0 {
		l.Trace().Msg("Calculating score")
		res.Score = float32(res.TopicNum-res.Min) / float32(res.Max-res.Min) * 100
	}
	l.Trace().Msg("Returning saved score")
	return res.Score
}

func (res *Result) String() string {
	if res == nil {
		return ""
	}
	return fmt.Sprintf("%s, %f, %d", res.Language, res.GetScore(), res.TopicNum)
}
