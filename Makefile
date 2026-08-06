.PHONY: swag-install swag run build

# Install CLI generator swag (sekali saja per environment/CI runner)
swag-install:
	go install github.com/swaggo/swag/cmd/swag@v1.16.4

# Generate/refresh docs/ dari anotasi @... di komentar handler.
# WAJIB dijalankan setiap kali anotasi swag berubah, dan WAJIB dijalankan
# sebelum build pertama kali (folder docs/ belum ada di git, hasil generate).
swag:
	swag init -g cmd/main.go -o docs

run: swag
	go run ./cmd/main.go

build: swag
	CGO_ENABLED=0 go build -ldflags="-s -w" -o gowms-backend ./cmd/main.go
