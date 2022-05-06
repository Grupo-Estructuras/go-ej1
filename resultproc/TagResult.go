package resultproc

import (
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
