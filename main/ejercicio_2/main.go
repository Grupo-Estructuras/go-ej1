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

	l.Trace().Msg("Scrape github")
	topics, err := sc.ScrapeInterest()
	if err != nil {
		l.Error().Err(err).Msg("Could not parse github!")
		return err
	}
	l.Trace().Msg("Creating result list")
	res := resultproc.CreateTagResultList(topics, app.Logger)
	l.Trace().Msg("Sorting results")
	res.TagSort()

	l.Trace().Msg("Print results")
	fmt.Printf(res.String())

	l.Trace().Msg("Creating graph")
	err = res.Graph(app.Config.HtmlFile)
	if err != nil {
		l.Error().Err(err).Msg("Could not graph results")
		return err
	}
	l.Trace().Msg("Open graph")
	app.OpenGraph()
	if err != nil {
		l.Error().Err(err).Msgf("Could not open graph. Please manually open %v in your browser.", app.Config.HtmlFile)
		return err
	}

	l.Trace().Msg("Exiting run without errors...")
	return nil
}
