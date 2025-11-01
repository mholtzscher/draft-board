package handlers

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

var layoutTemplate *template.Template

func init() {
	workDir, _ := os.Getwd()
	layoutPath := filepath.Join(workDir, "web", "templates", "layout.html")
	layoutTemplate = template.Must(template.ParseFiles(layoutPath))
}

func renderTemplate(w http.ResponseWriter, content string, title string) error {
	w.Header().Set("Content-Type", "text/html")
	data := map[string]interface{}{
		"Title":   title,
		"Content": template.HTML(content),
	}
	return layoutTemplate.Execute(w, data)
}

func renderError(w http.ResponseWriter, message string, title string) error {
	content := `
		<div class="bg-red-500/20 border border-red-500 rounded-lg p-6 mb-6">
			<h2 class="text-xl font-semibold text-red-500 mb-2">Error</h2>
			<p class="text-tokyo-night-fg-dim">` + template.HTMLEscapeString(message) + `</p>
		</div>
		<a href="/" class="inline-block mt-4 px-4 py-2 bg-tokyo-night-accent hover:bg-tokyo-night-accent-hover text-white rounded-lg">
			Go Home
		</a>
	`
	return renderTemplate(w, content, title)
}
