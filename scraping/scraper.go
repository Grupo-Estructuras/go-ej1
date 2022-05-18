package scraping

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"webscraping/common"

	"github.com/rs/zerolog"
)

type Scraper struct {
	Config *Scraperconfig
	Logger zerolog.Logger
}

type Scraperconfig struct {
	Tiobesiteformat      string            `json:"tiobe_site_format" yaml:"tiobe_site_format"`
	Githubsiteformat     string            `json:"github_site_format" yaml:"github_site_format"`
	Aliases              map[string]string `json:"aliases" yaml:"aliases"`
	RetryDelaysMs        []int             `json:"retry_delays_ms" yaml:"retry_delays_ms"`
	MaxPagesInterest     int               `json:"max_pages_interest" yaml:"max_pages_interest"`
	Interest             string            `json:"interest" yaml:"interest"`
	MaxParallel          int               `json:"max_parallel" yaml:"max_parallel"`
	Githubinterestformat string            `json:"github_interest_format" yaml:"github_interest_format"`
}

func GetDefaultScraperConfig(logger zerolog.Logger) Scraperconfig {
	l := logger.With().Str("function", "GetDefaultScraperConfig").Logger()
	l.Trace().Msg("Creando configuración por defecto.")
	return Scraperconfig{
		Tiobesiteformat:      "https://www.tiobe.com/tiobe-index/",
		Githubsiteformat:     "https://github.com/topics/%v",
		Aliases:              map[string]string{"C++": "cpp", "C#": "csharp", "Delphi/Object Pascal": "delphi", "Classic Visual Basic": "visual-basic"},
		RetryDelaysMs:        []int{300, 600, 1200},
		MaxPagesInterest:     10,
		Interest:             "sort",
		MaxParallel:          5,
		Githubinterestformat: "https://github.com/topics/%v?o=desc&page=%v",
	}
}

func (sc *Scraper) ScrapeTiobe() ([]string, error) {
	l := sc.Logger.With().Str("method", "ScraperTiobe").Logger()

	l.Trace().Str("url", sc.Config.Tiobesiteformat).Msgf("Accediendo a tiobe.")
	response, err := http.Get(sc.Config.Tiobesiteformat)
	if err != nil {
		l.Error().Err(err).Msg("No se pudo acceder a tiobe!")
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		for _, delay := range sc.Config.RetryDelaysMs {
			l.Warn().Int("Error code", response.StatusCode).Int("Tiempo espera", delay).Msg("La página retorno un error. Reintentando...")
			time.Sleep(time.Millisecond * time.Duration(delay))
			response, err = http.Get(sc.Config.Tiobesiteformat)
			if err != nil {
				l.Error().Err(err).Msg("No se pudo acceder a tiobe!")
				return nil, err
			}
			if response.StatusCode == http.StatusOK {
				break
			}
		}
	}
	if response.StatusCode != http.StatusOK {
		l.Error().Int("Código respuesta", response.StatusCode).Msg("No se pudo acceder en los intentos configurados!")
		return nil, common.NewStatusCodeError(response.StatusCode)
	}
	l.Trace().Str("url", sc.Config.Tiobesiteformat).Msgf("Leyendo contenido a cadena")
	content, err := io.ReadAll(response.Body)
	if err != nil {
		l.Error().Err(err).Msg("No se pudo leer todo el contenido!")
		return nil, err
	}
	l.Trace().Str("url", sc.Config.Tiobesiteformat).Msgf("Cerrando lector")
	response.Body.Close()

	l.Trace().Msgf("Compilando la expresión regular para la tabla de los top 20")
	rt := regexp.MustCompile(`<table.*id="top20".*>(.|\n)*?</table>`)
	l.Trace().Msgf("Buscando la tabla top 20")
	content = rt.Find(content)
	if content == nil {
		err := common.NewParseError("top 20 table")
		l.Error().Err(err).Msg("No se encontró la tabla!")
		return nil, err
	}

	l.Trace().Msgf("Compilando la expresión regular para buscar las filas de las tablas")
	rtd := regexp.MustCompile("<td.*?>.*?</td>")
	l.Trace().Msgf("Buscando todas las filas de la tabla")
	tabledata := rtd.FindAll(content, 140)
	if content == nil {
		err := common.NewParseError("table data")
		l.Error().Err(err).Msg("No se encontro contenido en la tabla!")
		return nil, err
	}

	l.Trace().Msgf("Compilando la expresión regular para reemplazar los tags  html")
	rtdr := regexp.MustCompile("</?td>")
	l.Trace().Msgf("Limpiando filas de contenido html")
	var languages []string
	for i := 4; i < 140; i += 7 {
		lang := string(rtdr.ReplaceAll(tabledata[i], []byte{}))
		l.Trace().Msgf("Agregando lenguaje %v", lang)
		languages = append(languages, lang)
	}

	l.Trace().Msgf("Pasar lista a función de alias")
	languages = sc.aliasreplace(languages)

	l.Trace().Msgf("EXIT")
	return languages, nil
}

func (sc *Scraper) aliasreplace(original []string) []string {
	l := sc.Logger.With().Str("method", "aliasreplace").Logger()

	l.Trace().Msg("Reemplazando lenguajes con sus alias")
	var replaced []string
	for _, lang := range original {
		replaceLang := sc.Config.Aliases[lang]
		if replaceLang != "" {
			l.Trace().Msgf("Reemplazando %v con alias %v", lang, replaceLang)
			replaced = append(replaced, replaceLang)
		} else {
			l.Trace().Msgf("%v hno tiene original, usando original", lang)
			replaced = append(replaced, lang)
		}
	}
	l.Trace().Msgf("EXIT")
	return replaced
}

func (sc *Scraper) ScrapeGithub(languages []string) (map[string]int32, error) {
	l := sc.Logger.With().Str("method", "ScrapeGithub").Logger()

	l.Trace().Msg("Preparando para scraping de github")
	ret := make(map[string]int32)
	l.Trace().Msgf("Compilando expresión regular para obtener línea con número de repositorios")
	rtopicLine := regexp.MustCompile(`Here\s+are\s+\d+(,\d*)*\s+public\s+repositories\s+matching\s+this\s+topic...`)
	l.Trace().Msgf("Compilando expresión regular para el número")
	rtopicnumber := regexp.MustCompile(`\d+(,\d*)*`)

	var lastError error
	var errMutex sync.Mutex
	var mapMutex sync.Mutex
	var wg sync.WaitGroup
	maxchannel := make(chan struct{}, sc.Config.MaxParallel)
	for _, lang := range languages {
		wg.Add(1)
		lang := lang

		go func() {
			defer wg.Done()
			// Contar, bloquea si se estan ejecutando ya MaxParallel rutinas
			maxchannel <- struct{}{}
			url := fmt.Sprintf(sc.Config.Githubsiteformat, lang)
			l.Trace().Str("url", url).Msgf("Haciendo consulta HTTP a github")
			response, err := http.Get(url)
			if err != nil {
				l.Error().Err(err).Msg("No se pudo acceder a Github, saltando...")
				errMutex.Lock()
				lastError = err
				errMutex.Unlock()
				<-maxchannel
				return
			}
			if response.StatusCode != http.StatusOK {
				for _, delay := range sc.Config.RetryDelaysMs {
					l.Warn().Int("Código error", response.StatusCode).Int("Tiempo espera", delay).Msg("Se retorno un error. Reintentando después de tiempo espera...")
					time.Sleep(time.Millisecond * time.Duration(delay))
					response, err = http.Get(url)
					if err != nil {
						l.Error().Err(err).Msg("No se pudo acceder a github!")
						errMutex.Lock()
						lastError = err
						errMutex.Unlock()
						<-maxchannel
						return
					}
					if response.StatusCode == http.StatusOK {
						break
					}
				}
			}
			if response.StatusCode != http.StatusOK {
				l.Error().Int("StatusCode", response.StatusCode).Msg("No se pudo acceder a github en intentos configurados!")
				errMutex.Lock()
				lastError = common.NewStatusCodeError(response.StatusCode)
				errMutex.Unlock()
				<-maxchannel
				return
			}
			l.Trace().Msg("Leer todo el contenido a cadena")
			content, err := io.ReadAll(response.Body)
			if err != nil {
				l.Error().Err(err).Msg("No se pudo leer todo, reintentando...")
				errMutex.Lock()
				lastError = err
				errMutex.Unlock()
				<-maxchannel
				return
			}
			l.Trace().Msg("Cerrando lector")
			response.Body.Close()
			l.Trace().Msg("Usando expresión regular de la línea de número")
			content = rtopicLine.Find(content)
			if content == nil {
				err := common.NewParseError("topic line")
				l.Error().Err(err).Msg("No se encontró lo buscado! Saltando...")
				errMutex.Lock()
				lastError = err
				errMutex.Unlock()
				<-maxchannel
				return
			}
			l.Trace().Msg("Buscando número con expresión regular")
			content = rtopicnumber.Find(content)
			if content == nil {
				err := common.NewParseError("topic number")
				l.Error().Err(err).Msg("No se encontró el número! Saltando...")
				errMutex.Lock()
				lastError = err
				errMutex.Unlock()
				<-maxchannel
				return
			}
			l.Trace().Msg("Leyendo número")
			num, err := strconv.ParseInt(strings.ReplaceAll(string(content), ",", ""), 10, 32)
			if err != nil {
				l.Error().Err(err).Msg("No se pudo convertir a número! Saltando topic...")
				errMutex.Lock()
				lastError = err
				errMutex.Unlock()
				<-maxchannel
				return
			}
			mapMutex.Lock()
			ret[lang] = int32(num)
			mapMutex.Unlock()
			<-maxchannel
		}()
	}
	wg.Wait()
	close(maxchannel)
	return ret, lastError
}

func (sc *Scraper) ScrapeInterest() (map[string]int, error) {
	l := sc.Logger.With().Str("method", "ScrapeGithub").Logger()

	l.Trace().Msgf("Preparando para scraping de github: %v", sc.Config.Interest)
	topics := make(map[string]int)

	l.Trace().Msgf("Preparando expresiónes regulares para tags: %v", sc.Config.Interest)
	l.Trace().Msgf("Compilando expresiónes regulares...")
	rarticle := regexp.MustCompile(`<article.*>(.|\n)*?</article>`)
	rtimehtml := regexp.MustCompile(`<relative-time.*>(.|\n)*?</relative-time>`)
	rtimestamp := regexp.MustCompile(`\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)
	rtag := regexp.MustCompile(`<a.*topic-tag topic-tag.*>(.|\n)*?</a>`)
	rtagbeg := regexp.MustCompile(`<a.*topic-tag topic-tag.*>`)
	rtagfin := regexp.MustCompile(`</a>`)

	l.Trace().Msgf("Leer tiempo referencia")
	now := time.Now()

	var lastError error
	var errMutex sync.Mutex
	var mapMutex sync.Mutex
	var wg sync.WaitGroup
	maxchannel := make(chan struct{}, sc.Config.MaxParallel)

	// Debemos empezar en pagina 1, porque si no github en pagina 0 y 1 retorna el mismo contenido
	for i := 1; i <= sc.Config.MaxPagesInterest; i++ {
		wg.Add(1)
		page := i

		go func() {
			defer wg.Done()
			// Contar, bloquea si se estan ejecutando ya MaxParallel rutinas
			maxchannel <- struct{}{}

			url := fmt.Sprintf(sc.Config.Githubinterestformat, strings.ToLower(sc.Config.Interest), page)
			l.Trace().Str("url\n", url).Msgf("Haciendo consulta HTTP a github")
			response, err := http.Get(url)
			if err != nil {
				l.Error().Err(err).Msg("No se pudo acceder a github! Saltando tag...")
				errMutex.Lock()
				lastError = err
				errMutex.Unlock()
				<-maxchannel
				return
			}
			if response.StatusCode != http.StatusOK {
				for _, delay := range sc.Config.RetryDelaysMs {
					l.Warn().Int("Error code", response.StatusCode).Int("Delay", delay).Msg("Página retorno error. Reintentando en...")
					time.Sleep(time.Millisecond * time.Duration(delay))
					response, err = http.Get(url)
					if err != nil {
						l.Error().Err(err).Msg("No se pudo acceder a github!")
						errMutex.Lock()
						lastError = err
						errMutex.Unlock()
						<-maxchannel
						return
					}
					if response.StatusCode == http.StatusOK {
						break
					}
				}
			}
			if response.StatusCode != http.StatusOK {
				l.Error().Int("StatusCode", response.StatusCode).Msg("No se pudo acceder a github en intentos configurados!")
				errMutex.Lock()
				lastError = common.NewStatusCodeError(response.StatusCode)
				errMutex.Unlock()
				<-maxchannel
				return
			}
			l.Trace().Msg("Leyendo todo el contenido a cadena")
			content, err := io.ReadAll(response.Body)
			if err != nil {
				l.Error().Err(err).Msg("No se pudo leer todo! Saltando topic...")
				errMutex.Lock()
				lastError = err
				errMutex.Unlock()
				<-maxchannel
				return
			}
			l.Trace().Msg("Cerrando lector")
			response.Body.Close()
			l.Trace().Msg("Usando expresión regular para encontrar artículo")
			articles := rarticle.FindAll(content, -1)

			l.Trace().Msg("Procesando artículos")
			for _, article := range articles {
				l.Trace().Msg("Buscar tiempo")
				timehtml := rtimehtml.Find(article)
				timebyte := rtimestamp.Find(timehtml)
				timestr := strings.ReplaceAll(string(timebyte), "\"", "")
				updtime, err := time.Parse(time.RFC3339, timestr)
				if timestr == "" {
					l.Trace().Msg("Saltando articulo sin tiempo (no es repositorio)")
					continue
				}
				if err != nil {
					l.Error().Err(err).Msg("Error leyendo tiempo, saltando página.")
					continue
				}
				l.Trace().Msg("Calculando diferencia en tiempo")
				if now.Sub(updtime) > time.Duration(30*24)*time.Hour {
					l.Trace().Msg("Este artículo es de hace más de 30 días, saltando...")
					continue
				}
				l.Trace().Msg("Diferencia menor a 30 días, procesando...")
				l.Trace().Msg("Usando expresión regular encontrar tags")
				tags := rtag.FindAll(article, -1)
				for _, tag := range tags {
					l.Trace().Msg("Procesando tag")
					tag = rtagbeg.ReplaceAll(tag, []byte{})
					tag = rtagfin.ReplaceAll(tag, []byte{})
					tagstr := string(tag)
					l.Trace().Msg("Cortando todo menos texto")
					tagstr = strings.TrimSpace(tagstr)

					mapMutex.Lock()
					topics[tagstr] = topics[tagstr] + 1
					mapMutex.Unlock()
				}
			}
			<-maxchannel
		}()
	}
	wg.Wait()
	return topics, lastError
}
