package site

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	texttemplate "text/template"
)

// TemplateManager manages HTML and Markdown templates
type TemplateManager struct {
	htmlTemplates     map[string]*template.Template
	markdownTemplates map[string]*texttemplate.Template
	templatesDir      string
	enabled           bool
}

// NavLink represents a navigation link
type NavLink struct {
	URL  string
	Text string
}

// NewTemplateManager creates a new template manager and loads all templates
func NewTemplateManager(templatesDir string) (*TemplateManager, error) {
	tm := &TemplateManager{
		htmlTemplates:     make(map[string]*template.Template),
		markdownTemplates: make(map[string]*texttemplate.Template),
		templatesDir:      templatesDir,
		enabled:           false,
	}

	// Check if templates directory exists
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		log.Printf("Templates directory %s does not exist, falling back to hardcoded templates", templatesDir)
		return tm, nil
	}

	tm.enabled = true

	// Load HTML templates
	htmlDir := filepath.Join(templatesDir, "html")
	htmlFiles := []string{"base.html", "index.html", "nav.html", "stat-detail.html", "subdir-index.html"}
	for _, file := range htmlFiles {
		path := filepath.Join(htmlDir, file)
		tmpl, err := template.ParseFiles(path)
		if err != nil {
			return nil, fmt.Errorf("failed to parse HTML template %s: %w", file, err)
		}
		name := file[:len(file)-5] // Remove .html extension
		tm.htmlTemplates[name] = tmpl
		log.Printf("Loaded HTML template: %s", name)
	}

	// Load Markdown templates
	mdDir := filepath.Join(templatesDir, "markdown")
	mdFiles := []string{"index.md", "stat-detail.md", "subdir-index.md"}
	for _, file := range mdFiles {
		path := filepath.Join(mdDir, file)
		tmpl, err := texttemplate.ParseFiles(path)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Markdown template %s: %w", file, err)
		}
		name := file[:len(file)-3] // Remove .md extension
		tm.markdownTemplates[name] = tmpl
		log.Printf("Loaded Markdown template: %s", name)
	}

	return tm, nil
}

// IsEnabled returns true if external templates are loaded
func (tm *TemplateManager) IsEnabled() bool {
	return tm.enabled
}

// RenderHTML executes an HTML template with the given data
func (tm *TemplateManager) RenderHTML(name string, data interface{}) (string, error) {
	tmpl, ok := tm.htmlTemplates[name]
	if !ok {
		return "", fmt.Errorf("HTML template %s not found", name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute HTML template %s: %w", name, err)
	}

	return buf.String(), nil
}

// RenderMarkdown executes a Markdown template with the given data
func (tm *TemplateManager) RenderMarkdown(name string, data interface{}) (string, error) {
	tmpl, ok := tm.markdownTemplates[name]
	if !ok {
		return "", fmt.Errorf("Markdown template %s not found", name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute Markdown template %s: %w", name, err)
	}

	return buf.String(), nil
}
