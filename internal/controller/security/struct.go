package security

import (
	"github.com/projsonal/gostock/pkg/botcheck"
	"github.com/projsonal/gostock/pkg/captcha"
)

type Controller struct {
	botSvc     *botcheck.Service
	captchaSvc *captcha.Service
}

func New(botSvc *botcheck.Service, captchaSvc *captcha.Service) *Controller {
	return &Controller{botSvc: botSvc, captchaSvc: captchaSvc}
}

type CheckRequest struct {
	BotToken string `json:"bot_token"`
}

type CheckResponse struct {
	Passed   bool               `json:"passed"`
	BotToken string             `json:"bot_token,omitempty"`
	Captcha  *captcha.Challenge `json:"captcha,omitempty"`
}

type SolveRequest struct {
	CaptchaToken  string `json:"captcha_token" validate:"required"`
	CaptchaAnswer string `json:"captcha_answer" validate:"required"`
}
