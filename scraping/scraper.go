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
	Tiobesiteformat  string            `json:"tiobe_site_format" yaml:"tiobe_site_format"`
	Githubsiteformat string            `json:"github_site_format" yaml:"github_site_format"`
	Aliases          map[string]string `json:"aliases" yaml:"aliases"`
	RetryDelaysMs    []int             `json:"retry_delays_ms" yaml:"retry_delays_ms"`
}

func GetDefaultScraperConfig(logger zerolog.Logger) Scraperconfig {
	l := logger.With().Str("function", "GetDefaultScraperConfig").Logger()
	l.Trace().Msg("Creating default config.")
	return Scraperconfig{
		Tiobesiteformat:  "https://www.tiobe.com/tiobe-index/",
		Githubsiteformat: "https://github.com/topics/%v",
		Aliases:          map[string]string{"C++": "cpp", "C#": "csharp", "Delphi/Object Pascal": "delphi", "Classic Visual Basic": "visual-basic"},
		RetryDelaysMs:    []int{300, 600, 1200},
	}
}

func (sc *Scraper) ScrapeTiobe() ([]string, error) {
	l := sc.Logger.With().Str("method", "ScraperTiobe").Logger()

	l.Trace().Str("url", sc.Config.Tiobesiteformat).Msgf("Making HTTP request to tiobe")
	response, err := http.Get(sc.Config.Tiobesiteformat)
	if err != nil {
		l.Error().Err(err).Msg("Could not access tiobe!")
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		for _, delay := range sc.Config.RetryDelaysMs {
			l.Warn().Int("Error code", response.StatusCode).Int("Delay", delay).Msg("Got error code. Retrying in...")
			time.Sleep(time.Millisecond * time.Duration(delay))
			response, err = http.Get(sc.Config.Tiobesiteformat)
			if err != nil {
				l.Error().Err(err).Msg("Could not access tiobe!")
				return nil, err
			}
			if response.StatusCode == http.StatusOK {
				break
			}
		}
	}
	if response.StatusCode != http.StatusOK {
		l.Error().Int("StatusCode", response.StatusCode).Msg("Could not access tiobe after retries!")
		return nil, common.NewStatusCodeError(response.StatusCode)
	}
	l.Trace().Str("url", sc.Config.Tiobesiteformat).Msgf("Reading entire body in to string")
	content, err := io.ReadAll(response.Body)
	if err != nil {
		l.Error().Err(err).Msg("Could not read all from body!")
		return nil, err
	}
	l.Trace().Str("url", sc.Config.Tiobesiteformat).Msgf("Closing reader")
	response.Body.Close()

	l.Trace().Msgf("Compiling regular expression for scanning top20 table")
	rt := regexp.MustCompile(`<table.*id="top20".*>(.|\n)*?</table>`)
	l.Trace().Msgf("Searching top20 table")
	content = rt.Find(content)
	if content == nil {
		err := common.NewParseError("top 20 table")
		l.Error().Err(err).Msg("Could not find table!")
		return nil, err
	}

	l.Trace().Msgf("Compiling regular expression for scanning tabledata elements")
	rtd := regexp.MustCompile("<td.*?>.*?</td>")
	l.Trace().Msgf("Searching all table data elements")
	tabledata := rtd.FindAll(content, 140)
	if content == nil {
		err := common.NewParseError("table data")
		l.Error().Err(err).Msg("Could not find table elements!")
		return nil, err
	}

	l.Trace().Msgf("Compiling regular expression for replacing html part")
	rtdr := regexp.MustCompile("</?td>")
	l.Trace().Msgf("Parsing all table data elements")
	var languages []string
	for i := 4; i < 140; i += 7 {
		lang := string(rtdr.ReplaceAll(tabledata[i], []byte{}))
		l.Trace().Msgf("Adding language %v", lang)
		languages = append(languages, lang)
	}

	l.Trace().Msgf("Passing list to replace function")
	languages = sc.aliasreplace(languages)

	l.Trace().Msgf("EXIT")
	return languages, nil
}

func (sc *Scraper) aliasreplace(original []string) []string {
	l := sc.Logger.With().Str("method", "aliasreplace").Logger()

	l.Trace().Msg("Replacing languages with aliases")
	var replaced []string
	for _, lang := range original {
		replaceLang := sc.Config.Aliases[lang]
		if replaceLang != "" {
			l.Trace().Msgf("Replacing %v with alias %v", lang, replaceLang)
			replaced = append(replaced, replaceLang)
		} else {
			l.Trace().Msgf("%v has no alias, using original.", lang)
			replaced = append(replaced, lang)
		}
	}
	l.Trace().Msgf("EXIT")
	return replaced
}

func (sc *Scraper) ScrapeGithub(languages []string) (map[string]int32, error) {
	l := sc.Logger.With().Str("method", "ScrapeGithub").Logger()

	l.Trace().Msg("Setting up for scraping github")
	ret := make(map[string]int32)
	l.Trace().Msgf("Compiling regular expression for getting line with number")
	rtopicLine := regexp.MustCompile(`Here\s+are\s+\d+(,\d*)*\s+public\s+repositories\s+matching\s+this\s+topic...`)
	l.Trace().Msgf("Compiling regular expression for getting topic number")
	rtopicnumber := regexp.MustCompile(`\d+(,\d*)*`)

	var lastError error
	var errMutex sync.Mutex
	var mapMutex sync.Mutex
	var wg sync.WaitGroup
	for _, lang := range languages {
		wg.Add(1)
		lang := lang

		go func() {
			defer wg.Done()
			url := fmt.Sprintf(sc.Config.Githubsiteformat, lang)
			l.Trace().Str("url", url).Msgf("Making HTTP request to github")
			response, err := http.Get(url)
			if err != nil {
				l.Error().Err(err).Msg("Could not access github! Skipping...")
				errMutex.Lock()
				lastError = err
				errMutex.Unlock()
				return
			}
			if response.StatusCode != http.StatusOK {
				for _, delay := range sc.Config.RetryDelaysMs {
					l.Warn().Int("Error code", response.StatusCode).Int("Delay", delay).Msg("Got error code. Retrying in...")
					time.Sleep(time.Millisecond * time.Duration(delay))
					response, err = http.Get(sc.Config.Tiobesiteformat)
					if err != nil {
						l.Error().Err(err).Msg("Could not access github!")
						errMutex.Lock()
						lastError = err
						errMutex.Unlock()
						return
					}
					if response.StatusCode == http.StatusOK {
						break
					}
				}
			}
			if response.StatusCode != http.StatusOK {
				l.Error().Int("StatusCode", response.StatusCode).Msg("Could not access github after retries!")
				errMutex.Lock()
				lastError = common.NewStatusCodeError(response.StatusCode)
				errMutex.Unlock()
				return
			}
			l.Trace().Msg("Reading all content in to string")
			content, err := io.ReadAll(response.Body)
			if err != nil {
				l.Error().Err(err).Msg("Could not read all from body! Skipping...")
				errMutex.Lock()
				lastError = err
				errMutex.Unlock()
				return
			}
			l.Trace().Msg("Closing reader")
			response.Body.Close()
			l.Trace().Msg("Regex find topic line")
			content = rtopicLine.Find(content)
			if content == nil {
				err := common.NewParseError("topic line")
				l.Error().Err(err).Msg("Could not find topic line! Skipping...")
				errMutex.Lock()
				lastError = err
				errMutex.Unlock()
				return
			}
			l.Trace().Msg("Regex find topic number")
			content = rtopicnumber.Find(content)
			if content == nil {
				err := common.NewParseError("topic number")
				l.Error().Err(err).Msg("Could not find topic number! Skipping...")
				errMutex.Lock()
				lastError = err
				errMutex.Unlock()
				return
			}
			l.Trace().Msg("Parsing number")
			num, err := strconv.ParseInt(strings.ReplaceAll(string(content), ",", ""), 10, 32)
			if err != nil {
				l.Error().Err(err).Msg("Could not parse number! Skipping...")
				errMutex.Lock()
				lastError = err
				errMutex.Unlock()
				return
			}
			mapMutex.Lock()
			ret[lang] = int32(num)
			mapMutex.Unlock()
		}()
	}
	wg.Wait()
	return ret, lastError
}
