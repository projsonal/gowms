package asset_port

import "github.com/projsonal/gowms/internal/model"

type Repository interface {
	// ListByAsset — semua port yang PERNAH punya baris untuk asset ini
	// (yang belum pernah dipakai tidak punya baris — lihat catatan di
	// model.AssetPort), diurutkan nomor port menaik.
	ListByAsset(assetID uint) ([]model.AssetPort, error)
	// FindPort — cari satu baris port spesifik (assetID, portNumber), nil
	// kalau belum pernah dipakai (belum ada barisnya).
	FindPort(assetID uint, portNumber int) (*model.AssetPort, error)
	// Upsert — buat baris baru kalau (AssetID, PortNumber) belum ada,
	// update kalau sudah ada. Dipakai handler "isi/ubah port".
	Upsert(p *model.AssetPort) error
	// Clear — kosongkan satu port (Status jadi "kosong", data pelanggan/
	// child dikosongkan) TANPA menghapus barisnya (riwayat tetap ada).
	Clear(assetID uint, portNumber int) error
	// CountTerisi — jumlah port berstatus "terisi" milik satu aset,
	// dipakai badge "X dari Y port terisi" di frontend.
	CountTerisi(assetID uint) (int64, error)
}
