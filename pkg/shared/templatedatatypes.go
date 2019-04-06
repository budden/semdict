package shared

import "html/template"

// GeneralTemplateParams are params for templates/general.html
type GeneralTemplateParams struct {
	Message string
}

// SenseViewParams are params for templates/senseview.html
type SenseViewParams struct {
	Id     int32
	Word   string
	Phrase template.HTML
}

// LoginFormParams are params for templates/loginform.html
type LoginFormParams struct {
	//CaptchaID string
}
