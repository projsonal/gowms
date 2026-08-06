// Package captcha mengekspos endpoint untuk mengambil soal CAPTCHA
// self-hosted (lihat pkg/captcha untuk implementasi generate/verify-nya).
package captcha

import "github.com/projsonal/gostock/pkg/captcha"

// Controller menangani endpoint HTTP untuk generate CAPTCHA.
type Controller struct {
	svc *captcha.Service
}

// New membuat instance Controller Captcha.
func New(svc *captcha.Service) *Controller {
	return &Controller{svc: svc}
}
