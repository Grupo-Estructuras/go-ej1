package main

import (
	"fmt"
	"time"
	"webscraping/app"
	"webscraping/resultproc"
	"webscraping/scraping"

	flag "github.com/spf13/pflag"
)

func main() {
	start := time.Now()
	var app app.Application

	loglevel := flag.StringP("loglevel", "l", "info", "Log level")
	app.ConfigFile = flag.StringP("configfile", "c", "resource/config/app.config", "Configuration file")
	flag.Parse()
	err := app.Configure(*loglevel)
	if err != nil {
		app.Logger.Err(err).Msg("Error configuring application. Shutting down...")
		return
	}
	l := app.Logger.With().Str("function", "main").Logger()
	l.Info().Msg("Application started!")
	l.Trace().Msg("Application configured without errors.")

	l.Trace().Msg("Running app")
	err = run(&app)
	if err != nil {
		l.Err(err).Msg("Error running application. Shutting down...")
		return
	}

	stop := time.Now()
	l.Info().Msgf("Completed in %v", stop.Sub(start))
}

func run(app *app.Application) error {
	l := app.Logger.With().Str("struct", "app").Str("method", "main").Logger()

	l.Trace().Msg("Creating scraper object")
	sc := scraping.Scraper{Config: &app.Config.Scraper, Logger: app.Logger.With().Str("struct", "scraper").Logger()}
	var listatiobe []string
	l.Trace().Msg("Checking config to see if static list is used")
	if !app.Config.UseFixedList {
		l.Trace().Msg("Scraping list from tiobe website")
		var err error
		listatiobe, err = sc.ScrapeTiobe()
		if err != nil {
			l.Error().Err(err).Msg("Could not parse tiobe!")
			return err
		}
	} else {
		l.Trace().Msg("Using static list")
		listatiobe = app.Config.LangList
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
	res := resultproc.CreateLanguageResultList(langData, app.Logger)
	l.Trace().Str("file", app.Config.ResultFile).Msg("Save results to file")
	res.Save(app.Config.ResultFile)
	l.Trace().Msg("Print results")
	res.ScoreSort()
	fmt.Print(res.String())

	res.NumSort()
	err = res.Graph(app.Config.HtmlFile)
	if err != nil {
		l.Error().Err(err).Msg("Could not graph results")
		return err
	}
	err = app.OpenGraph()
	if err != nil {
		l.Error().Err(err).Msgf("Could not open graph. Please manually open %v in your browser.", app.Config.HtmlFile)
		return err
	}
	return nil
}
