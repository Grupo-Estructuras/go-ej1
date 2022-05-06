package app

import (
	"os"
	"os/exec"
	"runtime"
	"webscraping/fileconfig"
	"webscraping/scraping"

	"github.com/rs/zerolog"
)

type Application struct {
	ConfigFile *string
	Logger     zerolog.Logger
	Config     ApplicationConfig
}

type ApplicationConfig struct {
	UseFixedList bool                   `json:"usar_lista_fija" yaml:"usar_lista_fija"`
	LangList     []string               `json:"lista_lenguajes" yaml:"lista_lenguajes"`
	Scraper      scraping.Scraperconfig `json:"scraper" yaml:"scraper"`
	HtmlFile     string                 `json:"archivo_html_grafo" yaml:"archivo_html_grafo"`
	ResultFile   string                 `json:"archivo_resultado" yaml:"archivo_resultado"`
}

func (app *Application) Configure(loglevelstr string) error {
	loglevel, err := zerolog.ParseLevel(loglevelstr)
	var l zerolog.Logger
	log_writer := zerolog.ConsoleWriter{Out: os.Stdout}
	app.Logger = zerolog.New(log_writer).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	if err != nil {
		l = app.Logger.Level(zerolog.InfoLevel).With().Str("struct", "app").Str("method", "configure").Logger()
		l.Info().Err(err).Msg("Could not parse loglevel. Using default info")
	} else {
		l = app.Logger.Level(loglevel).With().Str("struct", "app").Str("method", "configure").Logger()
		l.Trace().Msg("Initialized logger")
	}
	app.Logger = l

	l.Trace().Msg("Setting default config")
	app.Config.Scraper = scraping.GetDefaultScraperConfig(app.Logger)
	app.Config.LangList = []string{}
	app.Config.UseFixedList = false
	app.Config.HtmlFile = "resource/grafo.html"
	app.Config.ResultFile = "resource/resultado.txt"

	l.Trace().Msg("Creating new fileconfigstore")
	fs := fileconfig.NewFileConfigstore(l, *app.ConfigFile)
	l.Trace().Msg("Loading configuration from file")
	err = fs.Load(&app.Config)

	return err
}

func (app *Application) OpenGraph() error {
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open", app.Config.HtmlFile}
	case "windows":
		args = []string{"cmd", "/c", "start", app.Config.HtmlFile}
	default:
		args = []string{"xdg-open", app.Config.HtmlFile}
	}
	cmd := exec.Command(args[0], args[1:]...)
	err := cmd.Run()
	return err
}
