package stats

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-i2p/go-i2pcontrol"

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
	// DHT Statistics
	DHTEnabled           bool
	TotalRouters         int
	IPv4Routers          int
	IPv6Routers          int
	FloodfillRouters     int
	ReachableRouters     int
	NTCP2Routers         int
	SSU2Routers          int
	MeanEntropy          float64
	MeanAddressEntropy   float64
	LowEntropyIdentities int
	LowEntropyAddresses  int
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
		DHTEnabled:                       false,
		TotalRouters:                     0,
		IPv4Routers:                      0,
		IPv6Routers:                      0,
		FloodfillRouters:                 0,
		ReachableRouters:                 0,
		NTCP2Routers:                     0,
		SSU2Routers:                      0,
	}
}

func NewStats(dht *DHT) (Stats, error) {
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
	// Better handling of zero data - set to -1 to indicate no data
	if ExploratoryBuildTotal == 0 {
		ExploratoryBuildRejectedPercent = -1
		ExploratoryBuildSucceededPercent = -1
		ExploratoryBuildExpiredPercent = -1
	} else {
		ExploratoryBuildRejectedPercent = percent(ExploratoryBuildRejected, ExploratoryBuildTotal)
		ExploratoryBuildSucceededPercent = percent(ExploratoryBuildSucceeded, ExploratoryBuildTotal)
		ExploratoryBuildExpiredPercent = percent(ExploratoryBuildExpired, ExploratoryBuildTotal)
	}
	log.Println("ExploratoryBuildTotal:", ExploratoryBuildTotal)
	log.Println("ExploratoryBuildRejected:", ExploratoryBuildRejected)
	log.Println("ExploratoryBuildSucceeded:", ExploratoryBuildSucceeded)
	log.Println("ExploratoryBuildExpired:", ExploratoryBuildExpired)
	log.Println("ExploratoryBuildRejectedPercent:", ExploratoryBuildRejectedPercent)
	log.Println("ExploratoryBuildSucceededPercent:", ExploratoryBuildSucceededPercent)
	log.Println("ExploratoryBuildExpiredPercent:", ExploratoryBuildExpiredPercent)

	// Collect DHT statistics if provided
	var totalRouters, ipv4Routers, ipv6Routers, floodfillRouters, reachableRouters, ntcp2Routers, ssu2Routers, lowEntropyCount, lowAddressEntropyCount int
	var meanEntropy, meanAddressEntropy float64
	dhtEnabled := false
	if dht != nil {
		dhtEnabled = true
		totalRouters = dht.CountRouters()
		ipv4Routers = dht.CountIPv4Routers()
		ipv6Routers = dht.CountIPv6Routers()
		floodfillRouters = dht.CountFloodfills()
		reachableRouters = dht.CountReachableRouters()
		ntcp2Routers = dht.CountNTCP2Routers()
		ssu2Routers = dht.CountSSU2Routers()
		meanEntropy = dht.CalculateAverageIdentityEntropy()
		meanAddressEntropy = dht.CalculateAverageAddressEntropy()
		lowEntropyCount = dht.CountLowEntropyIdentities()
		lowAddressEntropyCount = dht.CountLowEntropyAddresses()
		log.Println("DHT Total Routers:", totalRouters)
		log.Println("DHT IPv4 Routers:", ipv4Routers)
		log.Println("DHT IPv6 Routers:", ipv6Routers)
		log.Println("DHT Floodfill Routers:", floodfillRouters)
		log.Println("DHT Reachable Routers:", reachableRouters)
		log.Println("DHT NTCP2 Routers:", ntcp2Routers)
		log.Println("DHT SSU2 Routers:", ssu2Routers)
		log.Println("DHT Low Identity Entropy (< 1/2 mean):", lowEntropyCount)
		log.Println("DHT Mean Identity Entropy:", meanEntropy)
		log.Println("DHT Low Address Entropy (< 1/2 mean):", lowAddressEntropyCount)
		log.Println("DHT Mean Address Entropy:", meanAddressEntropy)
	}

	return Stats{
		CollectedDate:                    time.Now(),
		ExploratoryBuildRejected:         ExploratoryBuildRejected,
		ExploratoryBuildSucceeded:        ExploratoryBuildSucceeded,
		ExploratoryBuildExpired:          ExploratoryBuildExpired,
		ExploratoryBuildRejectedPercent:  ExploratoryBuildRejectedPercent,
		ExploratoryBuildSucceededPercent: ExploratoryBuildSucceededPercent,
		ExploratoryBuildExpiredPercent:   ExploratoryBuildExpiredPercent,
		DHTEnabled:                       dhtEnabled,
		TotalRouters:                     totalRouters,
		IPv4Routers:                      ipv4Routers,
		IPv6Routers:                      ipv6Routers,
		FloodfillRouters:                 floodfillRouters,
		ReachableRouters:                 reachableRouters,
		NTCP2Routers:                     ntcp2Routers,
		SSU2Routers:                      ssu2Routers,
		MeanEntropy:                      meanEntropy,
		MeanAddressEntropy:               meanAddressEntropy,
		LowEntropyIdentities:             lowEntropyCount,
		LowEntropyAddresses:              lowAddressEntropyCount,
	}, nil
}

func percent(explSuccess, explTotal int) int {
	if explTotal == 0 {
		return 0
	}
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
	successPct := fmt.Sprintf("%d", s.ExploratoryBuildSucceededPercent)
	rejectPct := fmt.Sprintf("%d", s.ExploratoryBuildRejectedPercent)
	expiredPct := fmt.Sprintf("%d", s.ExploratoryBuildExpiredPercent)
	if s.ExploratoryBuildSucceededPercent == -1 {
		successPct = "N/A"
		rejectPct = "N/A"
		expiredPct = "N/A"
	}

	markdown := fmt.Sprintf("### Stats for: %s\n\n#### Build Statistics\n\n - Exploratory Build Success Percentage: %s\n - Exploratory Build Rejection Percentage: %s\n - Exploratory Build Expired Percentage: %s\n - Exploratory Build Success: %d\n - Exploratory Build Reject: %d\n - Exploratory Build Expired: %d\n",
		s.CollectedDate.String(),
		successPct,
		rejectPct,
		expiredPct,
		s.ExploratoryBuildSucceeded,
		s.ExploratoryBuildRejected,
		s.ExploratoryBuildExpired)

	if s.DHTEnabled {
		if s.TotalRouters > 0 {
			markdown += fmt.Sprintf("\n#### DHT Network Statistics\n\n - Total Routers: %d\n - IPv4 Routers: %d\n - IPv6 Routers: %d\n - Floodfill Routers: %d\n - Reachable Routers: %d\n - NTCP2 Routers: %d\n - SSU2 Routers: %d\n - Mean Identity Entropy: %.2f\n - Mean Address Entropy: %.2f\n - Low Identity Entropy (< 1/2 mean): %d\n - Low Address Entropy (< 1/2 mean): %d\n",
				s.TotalRouters,
				s.IPv4Routers,
				s.IPv6Routers,
				s.FloodfillRouters,
				s.ReachableRouters,
				s.NTCP2Routers,
				s.SSU2Routers,
				s.MeanEntropy,
				s.MeanAddressEntropy,
				s.LowEntropyIdentities,
				s.LowEntropyAddresses,
			)
		} else {
			markdown += "\n#### DHT Network Statistics\n\n - DHT collection enabled but no routers found in netDb\n"
		}
	}

	return markdown
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
	return os.WriteFile(p, []byte(statBytes), 0o644)
}

func LoadStats(jsonStr string) (Stats, error) {
	var stats Stats
	err := json.Unmarshal([]byte(jsonStr), &stats)
	if err != nil {
		return Stats{}, err
	}
	return stats, nil
}
