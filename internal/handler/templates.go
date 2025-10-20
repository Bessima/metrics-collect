package handler

import (
	"html/template"
)

func ParseAllTemplates() (tmpl *template.Template) {
	return template.Must(template.ParseGlob("templates/*.html"))
}
