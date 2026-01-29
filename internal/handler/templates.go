package handler

import (
	"html/template"
)

func ParseAllTemplates() (tmpl *template.Template) {
	templates, err := template.ParseGlob("templates/*.html")
	if templates != nil {
		return template.Must(templates, err)
	}
	return nil
}
