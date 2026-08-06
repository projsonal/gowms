package health

type Controller struct {
	checker *Checker
}

func NewController(checker *Checker) *Controller {
	return &Controller{checker: checker}
}
