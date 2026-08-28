package constant

const (
	StatusSerialTersedia  = "tersedia"
	StatusSerialTerpasang = "terpasang"
	StatusSerialRusak     = "rusak"
)

const (
	ErrSerialTidakDitemukan    = "unit dengan nomor seri tersebut tidak ditemukan"
	ErrSerialSudahDipakai      = "nomor seri sudah terdaftar di sistem (dipakai unit lain)"
	ErrSerialJumlahTidakSesuai = "jumlah nomor seri yang diisi tidak sama dengan qty item"
	ErrSerialDuplikatInput     = "ada nomor seri yang diisi berulang pada permintaan yang sama"
	ErrSerialTidakTersedia     = "nomor seri tidak berstatus tersedia di gudang asal (sudah terpasang/rusak, atau bukan di gudang ini)"
	ErrSerialBarangTidakSesuai = "nomor seri tidak terdaftar untuk barang yang dimaksud"
	ErrSerialBarangBukanSerial = "barang ini tidak ditandai sebagai barang ber-nomor-seri (aktifkan IsSerialized dulu di Kelola Barang)"
	ErrSerialBarangTidakAda    = "barang tidak ditemukan"
	ErrSerialGudangTidakAda    = "gudang tidak ditemukan"
)
