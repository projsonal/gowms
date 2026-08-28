package captcha

import "github.com/projsonal/gowms/pkg/captcha"

type Controller struct {
	svc *captcha.Service
}

func New(svc *captcha.Service) *Controller {
	return &Controller{svc: svc}
}
