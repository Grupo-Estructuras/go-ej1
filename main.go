package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"webscraping/fileconfig"
	"webscraping/resultproc"
	"webscraping/scraping"

	"github.com/rs/zerolog"
	flag "github.com/spf13/pflag"
)

type application struct {
	configFile *string
	logger     zerolog.Logger
	config     applicationConfig
}

type applicationConfig struct {
	UseFixedList bool                   `json:"usar_lista_fija" yaml:"usar_lista_fija"`
	LangList     []string               `json:"lista_lenguajes" yaml:"lista_lenguajes"`
	Scraper      scraping.Scraperconfig `json:"scraper" yaml:"scraper"`
	HtmlFile     string                 `json:"archivo_html_grafo" yaml:"archivo_html_grafo"`
	ResultFile   string                 `json:"archivo_resultado" yaml:"archivo_resultado"`
}

func main() {
	var app application

	loglevel := flag.StringP("loglevel", "l", "info", "Log level")
	app.configFile = flag.StringP("configfile", "c", "resource/config/app.config", "Configuration file")
	flag.Parse()
	err := app.configure(*loglevel)
	if err != nil {
		app.logger.Err(err).Msg("Error configuring application. Shutting down...")
		return
	}
	l := app.logger.With().Str("function", "main").Logger()
	l.Info().Msg("Application started!")
	l.Trace().Msg("Application configured without errors.")

	l.Trace().Msg("Running app")
	err = app.run()
	if err != nil {
		l.Err(err).Msg("Error running application. Shutting down...")
		return
	}

	l.Trace().Msg("Running shutdown")
	app.shutDown()
}

func (app *application) configure(loglevelstr string) error {
	loglevel, err := zerolog.ParseLevel(loglevelstr)
	var l zerolog.Logger
	log_writer := zerolog.ConsoleWriter{Out: os.Stdout}
	app.logger = zerolog.New(log_writer).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	if err != nil {
		l = app.logger.Level(zerolog.InfoLevel).With().Str("struct", "app").Str("method", "configure").Logger()
		l.Info().Err(err).Msg("Could not parse loglevel. Using default info")
	} else {
		l = app.logger.Level(loglevel).With().Str("struct", "app").Str("method", "configure").Logger()
		l.Trace().Msg("Initialized logger")
	}
	app.logger = l

	l.Trace().Msg("Setting default config")
	app.config.Scraper = scraping.GetDefaultScraperConfig(app.logger)
	app.config.LangList = []string{}
	app.config.UseFixedList = false
	app.config.HtmlFile = "resource/grafo.html"
	app.config.ResultFile = "resource/resultado.txt"

	l.Trace().Msg("Creating new fileconfigstore")
	fs := fileconfig.NewFileConfigstore(l, *app.configFile)
	l.Trace().Msg("Loading configuration from file")
	err = fs.Load(&app.config)

	return err
}

func (app *application) run() error {
	l := app.logger.With().Str("struct", "app").Str("method", "main").Logger()

	l.Trace().Msg("Creating scraper object")
	sc := scraping.Scraper{Config: &app.config.Scraper, Logger: app.logger.With().Str("struct", "scraper").Logger()}
	var listatiobe []string
	l.Trace().Msg("Checking config to see if static list is used")
	if !app.config.UseFixedList {
		l.Trace().Msg("Scraping list from tiobe website")
		var err error
		listatiobe, err = sc.ScrapeTiobe()
		if err != nil {
			l.Error().Err(err).Msg("Could not parse tiobe!")
			return err
		}
	} else {
		l.Trace().Msg("Using static list")
		listatiobe = app.config.LangList
	}
	l.Trace().Msg("Trying to scrape entry from github")
	langData, err := sc.ScrapeGithub(listatiobe)
	if err != nil {
		if len(langData) > 0 {
			l.Error().Err(err).Msgf("Could only process %d/20 languages! Please verify connection and aliases", len(langData))
		} else {
			l.Error().Err(err).Msg("Could not process any languages! Aborting...")
			return err
		}
	}

	l.Trace().Msg("Create result list")
	res := resultproc.CreateResultList(langData, app.logger)
	l.Trace().Str("file", app.config.ResultFile).Msg("Save results to file")
	res.Save(app.config.ResultFile)
	l.Trace().Msg("Print results")
	res.ScoreSort()
	fmt.Print(res.String())

	res.NumSort()
	err = res.Graph(app.config.HtmlFile)
	if err != nil {
		l.Error().Err(err).Msg("Could not graph results")
		return err
	}
	err = app.openGraph()
	if err != nil {
		l.Error().Err(err).Msgf("Could not open graph. Please manually open %v in your browser.", app.config.HtmlFile)
		return err
	}
	return nil
}

func (app *application) openGraph() error {
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open", app.config.HtmlFile}
	case "windows":
		args = []string{"cmd", "/c", "start", app.config.HtmlFile}
	default:
		args = []string{"xdg-open", app.config.HtmlFile}
	}
	cmd := exec.Command(args[0], args[1:]...)
	err := cmd.Run()
	return err
}

func (app *application) shutDown() {

}
