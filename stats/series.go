package stats

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type Series struct {
	Stats []Stats
}

func (s Series) JSONBytes() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

func (s Series) SaveSeries(seriesFile string) error {
	log.Println("saving stats as whole series")
	dir := filepath.Dir(seriesFile)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	mjson, err := s.JSONBytes()
	if err != nil {
		return err
	}
	return os.WriteFile(seriesFile, mjson, 0o644)
}

func (s Series) SaveStats(seriesDir string) error {
	log.Println("saving stats as individual files")
	for _, stat := range s.Stats {
		log.Println("Saving stat:", stat)
		if err := stat.SaveStat(seriesDir); err != nil {
			return err
		}
	}
	return nil
}

func (s Series) Markdown() (string, error) {
	// Try using external template if available
	if templateManager != nil && templateManager.IsEnabled() {
		// Create a struct with the series data and individual markdown strings
		type SeriesData struct {
			Stats []Stats
		}
		data := SeriesData{Stats: s.Stats}
		rendered, err := templateManager.RenderMarkdown("index", data)
		if err != nil {
			log.Printf("Error rendering template: %v, falling back to hardcoded", err)
		} else {
			return rendered, nil
		}
	}

	// Fallback to hardcoded template
	final := `Exploratory Build Stats Log
---------------------------
`
	for _, v := range s.Stats {
		final = fmt.Sprintf("%s\n%s", final, v.Markdown())
	}
	return fmt.Sprintf("%s\n", final), nil
}

func (s Series) mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	rendered := markdown.Render(doc, renderer)
	prefix := []byte(`<div class="stats multiple chart" id="exploratorystats">`)
	suffix := []byte(`</div>`)
	final := append(prefix, rendered...)
	final = append(final, suffix...)
	return final
}

func (s *Series) HTMLBytes() ([]byte, error) {
	md, err := s.Markdown()
	if err != nil {
		return nil, err
	}
	return s.mdToHTML([]byte(md)), nil
}

func (s *Series) HTML() string {
	b, err := s.HTMLBytes()
	if err != nil {
		return `<div class="error"><p>An error occurred</p><p>` + err.Error() + `</p></div>`
	}
	//	log.Println(string(b))
	return string(b)
}

func (s Series) JSONString() (string, error) {
	b, err := s.JSONBytes()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *Series) UpdateSeries(dht *DHT) error {
	stats, err := NewStats(dht)
	if err != nil {
		return err
	}
	s.Stats = append(s.Stats, stats)
	return nil
}

func NewSeries(dht *DHT) (Series, error) {
	stats, err := NewStats(dht)
	if err != nil {
		return Series{}, err
	}
	return Series{
		Stats: []Stats{stats},
	}, nil
}

func LoadSeries(jsonstr string) (Series, error) {
	var s Series
	jsonBytes, err := os.ReadFile(jsonstr)
	if err != nil {
		return s, err
	}
	err = json.Unmarshal(jsonBytes, &s)
	return s, err
}
