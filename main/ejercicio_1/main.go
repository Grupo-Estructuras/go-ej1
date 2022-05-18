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
		app.Logger.Err(err).Msg("Error configurando aplicacion. Terminando...")
		return
	}
	l := app.Logger.With().Str("function", "main").Logger()
	l.Info().Msg("Aplicacion lanzada!")
	l.Trace().Msg("Aplicacion configurada sin errores.")

	l.Trace().Msg("Corriendo app")
	err = run(&app)
	if err != nil {
		l.Err(err).Msg("Error corriendo app. Apagando...")
		return
	}

	stop := time.Now()
	l.Info().Msgf("Completando en %v", stop.Sub(start))
}

func run(app *app.Application) error {
	l := app.Logger.With().Str("struct", "app").Str("method", "main").Logger()

	l.Trace().Msg("Creando objeto scraper")
	sc := scraping.Scraper{Config: &app.Config.Scraper, Logger: app.Logger.With().Str("struct", "scraper").Logger()}
	var listatiobe []string
	l.Trace().Msg("Verificando configuracion para determinar si usar lista estatica")
	if !app.Config.UseFixedList {
		l.Trace().Msg("Scrapeando tiobe")
		var err error
		listatiobe, err = sc.ScrapeTiobe()
		if err != nil {
			l.Error().Err(err).Msg("Error scraping de tiobe!")
			return err
		}
	} else {
		l.Trace().Msg("Usando lista estatica")
		listatiobe = app.Config.LangList
	}
	l.Trace().Msg("Intentando scraping de github")
	langData, err := sc.ScrapeGithub(listatiobe)
	if err != nil {
		if len(langData) > 0 {
			l.Error().Err(err).Msgf("Solo se procesaron %d/20 lenguajes! Por favor verificar conexión y aliases", len(langData))
		} else {
			l.Error().Err(err).Msg("No se pudieron procesar lenguajes! Cancelando...")
			return err
		}
	}

	l.Trace().Msg("Crear lista resultados")
	res := resultproc.CreateLanguageResultList(langData, app.Logger)
	l.Trace().Str("file", app.Config.ResultFile).Msg("Guardar resultados en archivo")
	res.Save(app.Config.ResultFile)
	l.Trace().Msg("Imprimir resultados")
	res.ScoreSort()
	fmt.Print(res.String())

	res.NumSort()
	err = res.Graph(app.Config.HtmlFile)
	if err != nil {
		l.Error().Err(err).Msg("No se pudo graficar")
		return err
	}
	err = app.OpenGraph()
	if err != nil {
		l.Error().Err(err).Msgf("No se pudo abrir archivo. Puede intentar abrir manualmente %v en su navegador.", app.Config.HtmlFile)
		return err
	}
	return nil
}
