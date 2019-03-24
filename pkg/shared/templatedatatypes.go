package shared

import "html/template"

// GeneralTemplateParams are params for templates/general.html
type GeneralTemplateParams struct {
	Message string
}

// ArticleViewParams are params for templates/articleview.html
type ArticleViewParams struct {
	Word   string
	Phrase template.HTML
}
