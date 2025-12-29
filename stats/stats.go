package stats

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/eyedeekay/go-i2pcontrol"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
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

// Global template manager - will be set by main
var templateManager TemplateManager

// TemplateManager interface for stats package
type TemplateManager interface {
	IsEnabled() bool
	RenderMarkdown(name string, data interface{}) (string, error)
}

// SetTemplateManager sets the global template manager
func SetTemplateManager(tm TemplateManager) {
	templateManager = tm
}

type Stats struct {
	CollectedDate                    time.Time
	ExploratoryBuildRejected         int
	ExploratoryBuildSucceeded        int
	ExploratoryBuildExpired          int
	ExploratoryBuildRejectedPercent  int
	ExploratoryBuildSucceededPercent int
	ExploratoryBuildExpiredPercent   int
}

func ErrStat() Stats {
	return Stats{
		CollectedDate:                    time.Now(),
		ExploratoryBuildRejected:         0,
		ExploratoryBuildSucceeded:        0,
		ExploratoryBuildExpired:          0,
		ExploratoryBuildRejectedPercent:  0,
		ExploratoryBuildSucceededPercent: 0,
		ExploratoryBuildExpiredPercent:   0,
	}
}

func NewStats() (Stats, error) {
	i2pcontrol.Initialize("localhost", "7657", "jsonrpc")
	_, err := i2pcontrol.Authenticate("itoopie")
	if err != nil {
		return ErrStat(), err
	}
	ExploratoryBuildRejected, err := i2pcontrol.ExploratoryBuildReject()
	if err != nil {
		return ErrStat(), err
	}
	ExploratoryBuildSucceeded, err := i2pcontrol.ExploratoryBuildSuccess()
	if err != nil {
		return ErrStat(), err
	}
	ExploratoryBuildExpired, err := i2pcontrol.ExploratoryBuildExpire()
	if err != nil {
		return ErrStat(), err
	}
	ExploratoryBuildTotal := ExploratoryBuildRejected + ExploratoryBuildSucceeded + ExploratoryBuildExpired
	ExploratoryBuildRejectedPercent := 0
	ExploratoryBuildSucceededPercent := 0
	ExploratoryBuildExpiredPercent := 0
	if ExploratoryBuildTotal == 0 {
		ExploratoryBuildTotal = 1
	}
	log.Println("ExploratoryBuildTotal:", ExploratoryBuildTotal)
	log.Println("ExploratoryBuildRejected:", ExploratoryBuildRejected)
	log.Println("ExploratoryBuildSucceeded:", ExploratoryBuildSucceeded)
	log.Println("ExploratoryBuildExpired:", ExploratoryBuildExpired)
	ExploratoryBuildRejectedPercent = percent(ExploratoryBuildRejected, ExploratoryBuildTotal)
	ExploratoryBuildSucceededPercent = percent(ExploratoryBuildSucceeded, ExploratoryBuildTotal)
	ExploratoryBuildExpiredPercent = percent(ExploratoryBuildExpired, ExploratoryBuildTotal)
	log.Println("ExploratoryBuildRejectedPercent:", ExploratoryBuildRejectedPercent)
	log.Println("ExploratoryBuildSucceededPercent:", ExploratoryBuildSucceededPercent)
	log.Println("ExploratoryBuildExpiredPercent:", ExploratoryBuildExpiredPercent)

	return Stats{
		CollectedDate:                    time.Now(),
		ExploratoryBuildRejected:         ExploratoryBuildRejected,
		ExploratoryBuildSucceeded:        ExploratoryBuildSucceeded,
		ExploratoryBuildExpired:          ExploratoryBuildExpired,
		ExploratoryBuildRejectedPercent:  ExploratoryBuildRejectedPercent,
		ExploratoryBuildSucceededPercent: ExploratoryBuildSucceededPercent,
		ExploratoryBuildExpiredPercent:   ExploratoryBuildExpiredPercent,
	}, nil
}

func percent(explSuccess, explTotal int) int {
	return int(float64(explSuccess) / float64(explTotal) * 100)
}

func (s Stats) JSONString() (string, error) {
	b, err := s.JsonBytes()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s Stats) Markdown() string {
	// Try using external template if available
	if templateManager != nil && templateManager.IsEnabled() {
		rendered, err := templateManager.RenderMarkdown("stat-detail", s)
		if err != nil {
			log.Printf("Error rendering template: %v, falling back to hardcoded", err)
		} else {
			return rendered
		}
	}

	// Fallback to hardcoded template
	return fmt.Sprintf("### Stats for: %s\n\n - Exploratory Build Success Percentage: %d\n - Exploratory Build Rejection Percentage: %d\n - Exploratory Build Expired Percentage: %d\n - Exploratory Build Success: %d\n - Exploratory Build Reject: %d\n - Exploratory Build Expired: %d\n",
		s.CollectedDate.String(),
		s.ExploratoryBuildSucceededPercent,
		s.ExploratoryBuildRejectedPercent,
		s.ExploratoryBuildExpiredPercent,
		s.ExploratoryBuildSucceeded,
		s.ExploratoryBuildRejected,
		s.ExploratoryBuildExpired)
}

func (s Stats) HTMLBytes() []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(s.Markdown()))

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	rendered := markdown.Render(doc, renderer)
	prefix := []byte(`<div class="stats single measurement" id="` + s.CollectedDate.String() + `">`)
	suffix := []byte(`</div>`)
	final := append(prefix, rendered...)
	final = append(final, suffix...)
	return final
}

func (s Stats) HTML() string {
	return string(s.HTMLBytes())
}

func (s Stats) JsonBytes() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

var DateTime = "2006-01-02-15:04:05"

func (s Stats) SaveStat(jsonDir string) error {
	jsonBytes, err := s.JsonBytes()
	if err != nil {
		return err
	}
	var fsp []string
	fsp = append(fsp, jsonDir)
	fspb := strings.Split(s.CollectedDate.Format(DateTime), "-")
	fsp = append(fsp, fspb...)
	fsd := filepath.Dir(filepath.Join(fsp...))
	log.Println("fsd", fsd)
	if err := os.MkdirAll(fsd, 0o755); err != nil {
		return err
	}
	p := filepath.Join(fsp...) + ".json"
	log.Println("  p", p)
	return os.WriteFile(p, jsonBytes, 0o644)
}

func (s Stats) SaveHTML(jsonDir string) error {
	statBytes := s.HTML()
	var fsp []string
	fsp = append(fsp, jsonDir)
	fspb := strings.Split(s.CollectedDate.Format(DateTime), "-")
	fsp = append(fsp, fspb...)
	fsd := filepath.Dir(filepath.Join(fsp...))
	log.Println("fsd", fsd)
	if err := os.MkdirAll(fsd, 0o755); err != nil {
		return err
	}
	p := filepath.Join(fsp...) + ".html"
	log.Println("  p", p)
	return os.WriteFile(p, []byte(header+statBytes+footer), 0o644)
}

func (s Stats) SaveMarkdown(jsonDir string) error {
	statBytes := s.Markdown()
	var fsp []string
	fsp = append(fsp, jsonDir)
	fspb := strings.Split(s.CollectedDate.Format(DateTime), "-")
	fsp = append(fsp, fspb...)
	fsd := filepath.Dir(filepath.Join(fsp...))
	log.Println("fsd", fsd)
	if err := os.MkdirAll(fsd, 0o755); err != nil {
		return err
	}
	p := filepath.Join(fsp...) + ".md"
	log.Println("  p", p)
	return os.WriteFile(p, []byte(header+statBytes+footer), 0o644)
}

func LoadStats(jsonStr string) (Stats, error) {
	var stats Stats
	err := json.Unmarshal([]byte(jsonStr), &stats)
	if err != nil {
		return Stats{}, err
	}
	return stats, nil
}
