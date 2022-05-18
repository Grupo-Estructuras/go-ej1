package resultproc

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

type TagResult struct {
	Logger zerolog.Logger
	Num    int
	Tag    string
}

type TagSort []TagResult

func (a TagSort) Len() int           { return len(a) }
func (a TagSort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a TagSort) Less(i, j int) bool { return a[i].Num < a[j].Num }

func (res *TagResult) String() string {
	if res == nil {
		return ""
	}
	return fmt.Sprintf("%-30s: %d", res.Tag, res.Num)
}

func (res *TagResult) Save(file *os.File) error {
	l := res.Logger.With().Str("method", "Save").Str("lang", res.Tag).Logger()

	l.Trace().Msg("Intentando guardar resultado")
	_, err := file.WriteString(fmt.Sprintf("%v\n", res.Tag))
	if err != nil {
		l.Error().Err(err).Msg("No se pudo escribir en archivo")
		return err
	}
	l.Trace().Msg("EXIT")
	return nil
}
