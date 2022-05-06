package app

import (
	"webscraping/scraping"

	"github.com/rs/zerolog"
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
