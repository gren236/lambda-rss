package tg

import (
	"bytes"
	"html/template"
)

const htmlTemplate = `
<b>{{ .Title }}</b>
<i>by {{ .Author }}</i>

{{ .Description }}
<a href="{{ .Link }}">Read more</a>
`

var htmlTemplateParsed = template.Must(template.New("post").Parse(htmlTemplate))

type Post struct {
	Title       string
	Author      string
	Description string
	Link        string
}

func (p *Post) ToHTMLPost() (string, error) {
	var htmlPost bytes.Buffer
	err := htmlTemplateParsed.Execute(&htmlPost, p)
	if err != nil {
		return "", err
	}

	return htmlPost.String(), nil
}
