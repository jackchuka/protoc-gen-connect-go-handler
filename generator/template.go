package generator

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// templateCache holds parsed templates
var templateCache = make(map[string]*template.Template)

// renderTemplate renders a template with the given context
func renderTemplate(templateName string, ctx Context) (string, error) {
	tmpl, err := getTemplate(templateName)
	if err != nil {
		return "", fmt.Errorf("failed to get template %s: %w", templateName, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	return buf.String(), nil
}

// getTemplate retrieves a template from cache or loads it
func getTemplate(name string) (*template.Template, error) {
	if tmpl, exists := templateCache[name]; exists {
		return tmpl, nil
	}

	templatePath := fmt.Sprintf("templates/%s.tmpl", name)
	content, err := templateFS.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	tmpl, err := template.New(name).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	templateCache[name] = tmpl
	return tmpl, nil
}
