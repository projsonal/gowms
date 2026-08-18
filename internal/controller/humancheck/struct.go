package humancheck

import "github.com/projsonal/gowms/pkg/humancheck"

type Controller struct {
	svc *humancheck.Service
}

func New(svc *humancheck.Service) *Controller {
	return &Controller{svc: svc}
}

type IssueResponse struct {
	Token string `json:"human_check_token"`
}
