package site

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/eyedeekay/i2p-stats/stats"
)

var header = `<!DOCTYPE html>
<html>
	<head>
		<meta charset="utf-8">
		<title>i2p-stats</title>
	</head>
	<body>
`

var footer = `
	</body>
</html>`

type StatsSite struct {
	stats.Series
	StatsDirectory string
}

func (s *StatsSite) SeriesFile() string {
	seriesFile := filepath.Join(s.StatsDirectory, "series.json")
	log.Println("series file:", seriesFile)
	return seriesFile
}

func (s *StatsSite) OutputHomePage() error {
	index := filepath.Join(s.StatsDirectory, "index.html")
	htmlBytes := s.HTML()
	log.Println("Generating index:", index)
	return os.WriteFile(index, []byte(htmlBytes), 0o644)
}

func (s *StatsSite) OutputMarkdownHomePage() error {
	index := filepath.Join(s.StatsDirectory, "README.md")
	htmlBytes, err := s.Markdown()
	if err != nil {
		return err
	}
	log.Println("Generating index:", index)
	return os.WriteFile(index, []byte(htmlBytes), 0o644)
}

func (s *StatsSite) OutputPages() error {
	for _, stat := range s.Stats {
		log.Println("Saving stat html:", stat)
		if err := stat.SaveHTML(s.StatsDirectory); err != nil {
			return err
		}
	}
	return nil
}

func (s *StatsSite) OutputMarkdownPages() error {
	for _, stat := range s.Stats {
		log.Println("Saving stat html:", stat)
		if err := stat.SaveMarkdown(s.StatsDirectory); err != nil {
			return err
		}
	}
	return nil
}

func (s *StatsSite) listSubdirsWithFiles() []string {
	subdirs := make(map[string]string)
	log.Println("SUBDIR LISTING HERE")
	err := filepath.Walk(s.StatsDirectory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				htmlFiles, _ := filepath.Glob(filepath.Join(path, "*.html"))
				jsonFiles, _ := filepath.Glob(filepath.Join(path, "*.json"))
				if len(htmlFiles) > 0 && len(jsonFiles) > 0 {
					if len(htmlFiles) == 1 {
						for _, hfile := range htmlFiles {
							if hfile == "index.html" {
								return nil
							}
						}
					}
					subdirs[path] = path
				}
			}

			return nil
		})
	if err != nil {
		log.Println(err)
	}
	return maptoslice(subdirs)
}

func maptoslice(m map[string]string) []string {
	var s []string
	for k := range m {
		s = append(s, k)
	}
	return s
}

func (s StatsSite) GenerateNavSection() string {
	lsd := s.listSubdirsWithFiles()
	if lsd == nil || len(lsd) == 0 {
		return ""
	}
	lines := "\n"
	lines += `<div id="nav" class="navigation sitecomponent list">`
	lines += "<ul>\n"
	lines += fmt.Sprintf("    <li><a href=\"%s\">%s</a></li>\n", "/", "/")
	for _, subdir := range lsd {
		lines += fmt.Sprintf("    <li><a href=\"%s\">%s</a></li>\n", subdir, subdir)
	}
	lines += "</ul>"
	lines += "</div>\n"
	return lines
}

func (s StatsSite) GenerateIndexPages() error {
	log.Println("Generating indices")
	lsd := s.listSubdirsWithFiles()
	if lsd == nil || len(lsd) == 0 {
		return nil
	}
	for _, subdir := range lsd {
		lines := "\n"
		lines += `<div id="nav" class="navigation sitecomponent list">`
		lines += "<ul>\n"
		lines += fmt.Sprintf("    <li><a href=\"%s\">%s</a></li>\n", "/", "/")
		lines += s.sanitize(fmt.Sprintf("    <li><a href=\"/%s\">%s</a></li>\n", subdir, subdir))
		files, err := ioutil.ReadDir(subdir)
		if err != nil {
			return err
		}
		for _, f := range files {
			lines += fmt.Sprintf("    <li><a href=\"%s\">%s</a></li>\n", f.Name(), f.Name())
		}
		lines += "</ul>"
		lines += "</div>\n"
		page := s.sanitize(header + lines + footer)
		index := filepath.Join(subdir, "index.html")
		log.Println("Generating index:", index)
		if err := os.WriteFile(index, []byte(page), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func (s StatsSite) GenerateMarkdownIndexPages() error {
	log.Println("Generating indices")
	lsd := s.listSubdirsWithFiles()
	if lsd == nil || len(lsd) == 0 {
		return nil
	}
	for _, subdir := range lsd {
		lines := "\n"
		// lines += `<div id="nav" class="navigation sitecomponent list">`
		// lines += "<ul>\n"
		lines += fmt.Sprintf(" - [%s](%s)\n", "/", "/")
		lines += s.sanitize(fmt.Sprintf(" - [%s](%s)\n", subdir, subdir))
		files, err := ioutil.ReadDir(subdir)
		if err != nil {
			return err
		}
		for _, f := range files {
			lines += fmt.Sprintf(" - [%s](%s)\n", f.Name(), f.Name())
		}
		// lines += "</ul>"
		// lines += "</div>\n"
		page := s.sanitize(lines)
		index := filepath.Join(subdir, "README.md")
		log.Println("Generating index:", index)
		if err := os.WriteFile(index, []byte(page), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func NewStatsSite(statsDirectory string) (StatsSite, error) {
	absStatsDirectory, err := filepath.Abs(statsDirectory)
	if err != nil {
		log.Printf("error getting absolute path: %v", err)
		return StatsSite{}, err
	}

	s := StatsSite{
		StatsDirectory: absStatsDirectory,
		Series: stats.Series{
			Stats: []stats.Stats{},
		},
	}

	err = os.MkdirAll(s.StatsDirectory, 0o755)
	if err != nil {
		log.Printf("error creating stats directory: %v", err)
		return StatsSite{}, err
	}

	if fileExists(s.SeriesFile()) {
		log.Println("series file exists")
		if s.Series, err = stats.LoadSeries(s.SeriesFile()); err != nil {
			log.Printf("error loading stats series: %v", err)
			return StatsSite{}, err
		}
		if err := s.UpdateSeries(); err != nil {
			log.Printf("error updating stats series with new entry: %v", err)
			return StatsSite{}, err
		}
	} else {
		fmt.Println("series file does not exist, creating new series")
		s.Series, err = stats.NewSeries()
		if err != nil {
			log.Printf("error creating new stats series: %v", err)
			return StatsSite{}, err
		}
	}
	if err := s.SaveStats(absStatsDirectory); err != nil {
		return StatsSite{}, err
	}
	if err := s.SaveSeries(s.SeriesFile()); err != nil {
		return StatsSite{}, err
	}
	log.Println("created new stats series")
	return s, err
}

func (s *StatsSite) Markdown() (string, error) {
	return s.Series.Markdown()
}

func (s *StatsSite) HTML() string {
	body := s.Series.HTML()
	return s.sanitize(header + s.GenerateNavSection() + body + footer)
}

func (s *StatsSite) sanitize(sanitizee string) string {
	v := strings.Replace(strings.Replace(sanitizee, s.StatsDirectory, "", -1), "//", "", -1)
	return strings.TrimPrefix(v, "/")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		log.Println("file does not exist:", path)
		return false
	}
	if err != nil {
		log.Fatal("unexpected startup error:", err)
		return false
	}
	log.Printf("file %s exists\n", path)
	return true
}
