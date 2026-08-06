package laporan

import (
	barangRepo "github.com/projsonal/gostock/internal/repositories/barang"
	barangKeluarRepo "github.com/projsonal/gostock/internal/repositories/barang_keluar"
	barangMasukRepo "github.com/projsonal/gostock/internal/repositories/barang_masuk"
	purchaseOrderRepo "github.com/projsonal/gostock/internal/repositories/po"
	"github.com/projsonal/gostock/internal/repositories/role"
	stockOpnameRepo "github.com/projsonal/gostock/internal/repositories/stockOpname"
	"github.com/projsonal/gostock/pkg/utils"
)

const exportRowLimit = 20000

type Controller struct {
	barangRepo       barangRepo.Repository
	poRepo           purchaseOrderRepo.Repository
	barangMasukRepo  barangMasukRepo.Repository
	barangKeluarRepo barangKeluarRepo.Repository
	stockOpnameRepo  stockOpnameRepo.Repository
	roleRepo         role.Repository
	jwtSvc           *utils.JWTService
}

func New(barangRepo barangRepo.Repository, poRepo purchaseOrderRepo.Repository, barangMasukRepo barangMasukRepo.Repository,
	barangKeluarRepo barangKeluarRepo.Repository, stockOpnameRepo stockOpnameRepo.Repository,
	roleRepo role.Repository, jwtSvc *utils.JWTService) *Controller {
	return &Controller{
		barangRepo: barangRepo, poRepo: poRepo, barangMasukRepo: barangMasukRepo,
		barangKeluarRepo: barangKeluarRepo, stockOpnameRepo: stockOpnameRepo,
		roleRepo: roleRepo, jwtSvc: jwtSvc,
	}
}

func bigPagination() utils.PaginationParams {
	return utils.PaginationParams{Page: 1, Limit: exportRowLimit}
}
