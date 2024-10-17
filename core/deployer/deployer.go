package deployer

import (
	"bytes"
	"html/template"
	"lamprey/core/db"
)

type Deployer interface {
	DeployArticle(page db.Page) error
	DeployData(page db.Page) error
}

type DummyDeployer struct{}

func (d DummyDeployer) DeployArticle(page db.Page) error {
	return nil
}

func (d DummyDeployer) DeployData(page db.Page) error {
	return nil
}

func GetPageHtml(page db.Page) (string, error) {
	templates, err := template.ParseFiles("views/layout.html.gotmpl")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("content").Parse("{{ define \"content\" }}" + page.Content + "{{ end }}")
	if err != nil {
		return "", err
	}

	templates.AddParseTree("content", tmpl.Tree)

	// String writer
	buf := bytes.NewBufferString("")

	// Render the template with the provided Page data
	err = templates.ExecuteTemplate(buf, "layout.html.gotmpl", page)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
