package laporan

import (
	barangRepo "github.com/projsonal/gowms/internal/repositories/barang"
	barangKeluarRepo "github.com/projsonal/gowms/internal/repositories/barang_keluar"
	barangMasukRepo "github.com/projsonal/gowms/internal/repositories/barang_masuk"
	barangRusakRepo "github.com/projsonal/gowms/internal/repositories/barang_rusak"
	purchaseOrderRepo "github.com/projsonal/gowms/internal/repositories/po"
	"github.com/projsonal/gowms/internal/repositories/role"
	stockOpnameRepo "github.com/projsonal/gowms/internal/repositories/stockOpname"
	"github.com/projsonal/gowms/pkg/utils"
)

const exportRowLimit = 20000

type Controller struct {
	barangRepo       barangRepo.Repository
	poRepo           purchaseOrderRepo.Repository
	barangMasukRepo  barangMasukRepo.Repository
	barangKeluarRepo barangKeluarRepo.Repository
	stockOpnameRepo  stockOpnameRepo.Repository
	barangRusakRepo  barangRusakRepo.Repository
	roleRepo         role.Repository
	jwtSvc           *utils.JWTService
}

func New(barangRepo barangRepo.Repository, poRepo purchaseOrderRepo.Repository, barangMasukRepo barangMasukRepo.Repository,
	barangKeluarRepo barangKeluarRepo.Repository, stockOpnameRepo stockOpnameRepo.Repository, barangRusakRepo barangRusakRepo.Repository,
	roleRepo role.Repository, jwtSvc *utils.JWTService) *Controller {
	return &Controller{
		barangRepo: barangRepo, poRepo: poRepo, barangMasukRepo: barangMasukRepo,
		barangKeluarRepo: barangKeluarRepo, stockOpnameRepo: stockOpnameRepo, barangRusakRepo: barangRusakRepo,
		roleRepo: roleRepo, jwtSvc: jwtSvc,
	}
}

func bigPagination() utils.PaginationParams {
	return utils.PaginationParams{Page: 1, Limit: exportRowLimit}
}
