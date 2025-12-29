package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/eyedeekay/i2p-stats/site"
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

func main() {
	flag.Parse()
	os.Chdir(*runDir)
	if statsite, err := site.NewStatsSite(*runDir); err != nil {
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
		if gitIsInstalled() {
			if gitDirExists(*runDir) {
				gitAddCmd := exec.Command("git", "add", statsite.StatsDirectory)
				gitAddCmd.Stdout = os.Stdout
				gitAddCmd.Stderr = os.Stderr
				gitAddCmd.Run()
			}
		}
	}
}

func appIsInstalled(app string) bool {
	_, err := exec.LookPath(app)
	if err != nil {
		gopath := os.Getenv("GOPATH")
		if gopath != "" {
			binPath := filepath.Join(gopath, "bin", app)
			_, err = os.Stat(binPath)
			if err == nil {
				log.Println("found", binPath)
				return true
			}
		}
		return false
	}
	log.Println("found", app)
	return true
}

func gitIsInstalled() bool {
	return appIsInstalled("git")
}

func gitDirExists(statsDir string) bool {
	log.Println("checking if", filepath.Join(statsDir, ".git"), "is a git directory")
	_, err := os.Stat(filepath.Join(statsDir, ".git"))
	if err == nil {
		log.Println("git dir exists")
	} else {
		log.Println("git dir does not exist")
	}
	return err == nil
}
