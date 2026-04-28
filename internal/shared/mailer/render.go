package mailer

import (
	"bytes"
	"html/template"
)

func RenderTemplate(name, rawTemplate string, data any) (string, error) {
	tmpl, err := template.New(name).Parse(rawTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
