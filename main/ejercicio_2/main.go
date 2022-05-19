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

	l.Trace().Msg("Scrapeando github")
	topics, err := sc.ScrapeInterest()
	if err != nil {
		l.Error().Err(err).Msg("No se pudo scrapear github!")
		return err
	}
	l.Trace().Msg("Creando lista resultado")
	res := resultproc.CreateTagResultList(topics, app.Logger)
	l.Trace().Msg("Ordenando resultados")
	res.TagSort()

	l.Trace().Msg("Guardar resultados")
	res.Save(app.Config.ResultFile)

	l.Trace().Msg("Imprimir resultados")
	fmt.Printf(res.String())

	l.Trace().Msg("Creando gráfica")
	err = res.Graph(app.Config.HtmlFile)
	if err != nil {
		l.Error().Err(err).Msg("No se pudo graficar")
		return err
	}
	l.Trace().Msg("Abriendo archivo grafica")
	app.OpenGraph()
	if err != nil {
		fmt.Printf("Para visualizar el resultado abre %v en su navegador.", app.Config.HtmlFile)
	}

	l.Trace().Msg("Saliendo sin errores...")
	return nil
}
