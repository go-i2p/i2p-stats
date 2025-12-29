package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/eyedeekay/i2p-stats/git"
	"github.com/eyedeekay/i2p-stats/site"
	"github.com/eyedeekay/i2p-stats/stats"
)

var Docroot = docroot

func docroot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	i2p := filepath.Join(home, "i2p")
	eepsite := filepath.Join(i2p, "eepsite")
	docroot := filepath.Join(eepsite, "docroot")
	weather := filepath.Join(docroot, "weather")
	os.MkdirAll(weather, 0o755)
	return weather
}

// Get the user's home directory.
// Build the path to the i2p directory inside the home directory.
// Build the path to the eepsite directory inside the i2p directory.
// Build the path to the docroot directory inside the eepsite directory.
// Return the docroot path.

var runDir = flag.String("dir", Docroot(), "directory to run from")
var templatesDir = flag.String("templates", "./templates", "directory containing templates")
var noGit = flag.Bool("no-git", false, "skip git operations")

func main() {
	flag.Parse()
	os.Chdir(*runDir)

	// Initialize template manager
	templateManager, err := site.NewTemplateManager(*templatesDir)
	if err != nil {
		log.Fatal(err)
	}

	// Set template manager for stats package
	stats.SetTemplateManager(templateManager)

	if statsite, err := site.NewStatsSite(*runDir, templateManager); err != nil {
		log.Fatal(err)
	} else {
		if err := statsite.OutputPages(); err != nil {
			log.Fatal(err)
		}
		if err := statsite.GenerateIndexPages(); err != nil {
			log.Fatal(err)
		}
		if err := statsite.OutputHomePage(); err != nil {
			log.Fatal(err)
		}
		if !*noGit {
			git.AddChanges(*runDir, statsite.StatsDirectory)
		}
	}
}
