package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/go-i2p/i2p-stats/git"
	"github.com/go-i2p/i2p-stats/site"
	"github.com/go-i2p/i2p-stats/stats"
	"github.com/go-i2p/logger"
)

var log = logger.GetGoI2PLogger()

var Docroot = docroot

func docroot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	i2p := filepath.Join(home, ".i2p")
	eepsite := filepath.Join(i2p, "eepsite")
	docroot := filepath.Join(eepsite, "docroot")
	weather := filepath.Join(docroot, "weather")
	os.MkdirAll(weather, 0o755)
	return weather
}

func netdbroot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	i2p := filepath.Join(home, ".i2p")
	netdb := filepath.Join(i2p, "netDb")
	return netdb
}

// Get the user's home directory.
// Build the path to the i2p directory inside the home directory.
// Build the path to the eepsite directory inside the i2p directory.
// Build the path to the docroot directory inside the eepsite directory.
// Return the docroot path.

var (
	runDir       = flag.String("dir", Docroot(), "directory to run from")
	templatesDir = flag.String("templates", "./templates", "directory containing templates")
	noGit        = flag.Bool("no-git", false, "skip git operations")
	dhtDir       = flag.String("dht", netdbroot(), "directory containing DHT RouterInfo files")
	geoipDB      = flag.String("geoip", "", "path to MaxMind GeoIP2 database file")
)

func main() {
	flag.Parse()
	os.Chdir(*runDir)

	// Initialize template manager
	templateManager, err := site.NewTemplateManager(*templatesDir)
	if err != nil {
		log.Printf("Warning: Template initialization failed: %v. Falling back to hardcoded templates.", err)
		templateManager = site.NewDisabledTemplateManager()
	}

	// Set template manager for stats package
	stats.SetTemplateManager(templateManager)

	// Initialize DHT if directory provided
	var dht *stats.DHT
	if *dhtDir != "" {
		log.Println("Initializing DHT analysis from:", *dhtDir)
		d, err := stats.NewDHT(*dhtDir)
		if err != nil {
			log.Printf("Warning: DHT initialization failed: %v", err)
		} else {
			dht = &d
		}
	}

	// Initialize GeoIP if database provided
	var geoip *stats.GeoIP
	if *geoipDB != "" {
		log.Println("Initializing GeoIP from:", *geoipDB)
		g, err := stats.NewGeoIP(*geoipDB)
		if err != nil {
			log.Printf("Warning: GeoIP initialization failed: %v", err)
		} else {
			geoip = g
			defer geoip.Close()
		}
	}

	if statsite, err := site.NewStatsSite(*runDir, templateManager, dht); err != nil {
		log.Fatal(err)
	} else {
		// Generate HTML output
		if err := statsite.OutputPages(); err != nil {
			log.Fatal(err)
		}
		if err := statsite.GenerateIndexPages(); err != nil {
			log.Fatal(err)
		}
		if err := statsite.OutputHomePage(); err != nil {
			log.Fatal(err)
		}
		// Generate Markdown output
		if err := statsite.OutputMarkdownPages(); err != nil {
			log.Fatal(err)
		}
		if err := statsite.OutputMarkdownHomePage(); err != nil {
			log.Fatal(err)
		}
		if err := statsite.GenerateMarkdownIndexPages(); err != nil {
			log.Fatal(err)
		}
		if !*noGit {
			git.AddChanges(*runDir, statsite.StatsDirectory)
		}
	}
}
