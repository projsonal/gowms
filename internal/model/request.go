package model

type IDParam struct {
	ID uint `params:"id" validate:"required"`
}

type PaginationQuery struct {
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
	Search string `query:"search"`
	Sort   string `query:"sort"`
}
