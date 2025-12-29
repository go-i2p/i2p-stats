# Migration Plan: External HTML and Markdown Templates

## Current State
- HTML/Markdown generation is hardcoded in Go code
- Header/footer HTML strings embedded in `site/site.go`
- Stats formatting logic embedded in `stats/stats.go` and `stats/series.go`
- Navigation and index generation built with string concatenation

## Goal
Minimum viable external template system for maintainability and flexibility.

## Implementation Steps

### 1. Create Template Directory Structure
```
templates/
├── html/
│   ├── base.html          # Header + footer wrapper
│   ├── index.html         # Homepage listing
│   ├── nav.html           # Navigation component
│   ├── stat-detail.html   # Single stat display
│   └── subdir-index.html  # Subdirectory index
└── markdown/
    ├── index.md           # Homepage listing
    ├── stat-detail.md     # Single stat display
    └── subdir-index.md    # Subdirectory index
```

### 2. Refactor Template Loading (`site/templates.go`)
- Create `TemplateManager` struct
- Load templates once at startup using `html/template` and `text/template`
- Cache parsed templates in memory
- Add error handling for missing templates

### 3. Update Stats Generation (`stats/stats.go`, `stats/series.go`)
- Replace hardcoded `Markdown()` methods to use external templates
- Replace `HTML()` methods to use templates (keep markdown-to-html conversion)
- Pass data structs to templates instead of building strings

### 4. Update Site Generation (`site/site.go`)
- Remove `header` and `footer` variables
- Update `HTML()` to use base template
- Update `GenerateNavSection()` to use nav template
- Update `GenerateIndexPages()` to use index templates
- Pass `TemplateManager` to `StatsSite` struct

### 5. Update Main (`main.go`)
- Initialize `TemplateManager` with template directory path
- Add `-templates` flag for template directory location (default: `./templates`)

## Minimal Template Examples

### HTML Base (`templates/html/base.html`)
```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>{{.Title}}</title>
</head>
<body>
    {{.Nav}}
    {{.Content}}
</body>
</html>
```

### Stat Detail (`templates/markdown/stat-detail.md`)
```markdown
### Stats for: {{.CollectedDate}}

- Exploratory Build Success Percentage: {{.ExploratoryBuildSucceededPercent}}
- Exploratory Build Rejection Percentage: {{.ExploratoryBuildRejectedPercent}}
- Exploratory Build Expired Percentage: {{.ExploratoryBuildExpiredPercent}}
```

## Testing Strategy
1. Create templates directory with all required files
2. Test template loading/parsing on startup
3. Verify output matches current implementation
4. No functional changes to output content

## Rollback
Keep current implementation as fallback if templates directory doesn't exist.

## Out of Scope (Future)
- CSS/styling enhancements
- JavaScript functionality  
- Advanced template features (partials, inheritance)
- Dynamic configuration of template locations
