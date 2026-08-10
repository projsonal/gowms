module github.com/projsonal/gowms

go 1.26.5

// -----------------------------------------------------------------------
// Dependensi untuk pkg/wa/whatsmeow_sender.go & cmd/whatsapp-pair BELUM
// otomatis ter-resolve di sini (tidak ada akses jaringan Go module proxy
// saat file ini ditulis). SEBELUM `go build`/`go run`, jalankan:
//
//	go get go.mau.fi/whatsmeow@latest
//	go get github.com/mdp/qrterminal/v3@latest
//	go get modernc.org/sqlite@latest
//	go mod tidy
//
// Ini hanya WAJIB kalau kamu memakai WHATSAPP_DRIVER=whatsmeow. Kalau
// tetap pakai driver "gateway" (default, HTTP ke gateway berbayar), baris
// di atas tidak perlu dijalankan — tapi pkg/wa/whatsmeow_sender.go tetap
// akan gagal compile karena import-nya belum ada di go.sum. Kalau tidak
// berniat pakai whatsmeow sama sekali, hapus saja file
// pkg/wa/whatsmeow_sender.go dan cmd/whatsapp-pair/main.go.
// -----------------------------------------------------------------------

require (
	github.com/go-playground/validator/v10 v10.30.3
	github.com/gofiber/fiber/v2 v2.52.14
	github.com/gofiber/swagger v1.1.1
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/joho/godotenv v1.5.1
	github.com/jung-kurt/gofpdf v1.16.2
	github.com/mdp/qrterminal/v3 v3.2.1
	github.com/pquerna/otp v1.5.0
	github.com/xuri/excelize/v2 v2.11.0
	go.mau.fi/whatsmeow v0.0.0-20260806224404-e277b766ab33
	golang.org/x/crypto v0.54.0
	google.golang.org/protobuf v1.36.11
	gorm.io/driver/postgres v1.6.2
	gorm.io/gorm v1.31.2
	modernc.org/sqlite v1.56.0
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/beeper/argo-go v1.1.2 // indirect
	github.com/coder/websocket v1.8.15 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/elliotchance/orderedmap/v3 v3.1.0 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/petermattis/goid v0.0.0-20260713124913-97594f28f5ca // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rs/zerolog v1.35.1 // indirect
	github.com/vektah/gqlparser/v2 v2.5.27 // indirect
	go.mau.fi/libsignal v0.2.2 // indirect
	go.mau.fi/util v0.9.12-0.20260717235539-f9ffa7eca58d // indirect
	golang.org/x/exp v0.0.0-20260709172345-9ea1abe57597 // indirect
	golang.org/x/term v0.45.0 // indirect
	modernc.org/libc v1.74.4 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	rsc.io/qr v0.2.0 // indirect
)

require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/gabriel-vasile/mimetype v1.4.13 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/spec v0.20.4 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.10.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/richardlehane/mscfb v1.0.7 // indirect
	github.com/richardlehane/msoleps v1.0.6 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.16.0 // indirect
	github.com/swaggo/files/v2 v2.0.2 // indirect
	github.com/swaggo/swag v1.16.4
	github.com/tiendc/go-deepcopy v1.7.2 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.51.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	github.com/xuri/efp v0.0.1 // indirect
	github.com/xuri/nfp v0.0.2-0.20250530014748-2ddeb826f9a9 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	golang.org/x/tools v0.48.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
