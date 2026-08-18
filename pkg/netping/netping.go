// Package netping mengecek konektivitas perangkat jaringan (ODC/ONT/ODP/OLT
// dsb) lewat ICMP ping, dipakai fitur tracking aset "label RSD berdasarkan
// ping" (lihat internal/controller/asset_gudang Ping/PingAll).
//
// Implementasi memanggil binary `ping` sistem (tersedia default di semua
// distro Linux) lewat os/exec, alih-alih membuka raw socket ICMP langsung
// dari Go (butuh privilese CAP_NET_RAW/root yang seringkali tidak tersedia
// di lingkungan hosting/container biasa). Trade-off ini didokumentasikan
// eksplisit supaya tidak dikira bug kalau server deploy tidak punya binary
// `ping` — pada kasus itu Check selalu mengembalikan error.
package netping

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"time"
)

var ErrInvalidTarget = errors.New("alamat IP aset tidak valid")

// Result — hasil satu kali pengecekan ping ke sebuah alamat IP.
type Result struct {
	Online bool
	// RTT nol kalau Online == false atau tidak berhasil di-parse.
	RTT time.Duration
}

// Check mengirim satu paket ICMP echo ke target dengan batas waktu timeout.
// Mengembalikan error hanya untuk kegagalan input/eksekusi (IP tidak valid,
// binary `ping` tidak ditemukan) — target yang tidak merespon BUKAN error,
// melainkan Result{Online: false}.
func Check(target string, timeout time.Duration) (Result, error) {
	if net.ParseIP(target) == nil {
		return Result{}, ErrInvalidTarget
	}
	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout+time.Second)
	defer cancel()

	waitSecs := strconv.Itoa(int(timeout.Seconds()))
	if waitSecs == "0" {
		waitSecs = "1"
	}
	// -c 1: satu paket saja. -W: batas waktu tunggu balasan (detik, GNU
	// iputils). Kompatibel dengan `ping` bawaan mayoritas distro Linux.
	cmd := exec.CommandContext(ctx, "ping", "-c", "1", "-W", waitSecs, target)
	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)

	if ctx.Err() != nil {
		return Result{}, fmt.Errorf("ping ke %s timeout: %w", target, ctx.Err())
	}
	if err != nil {
		// Exit code non-zero dari `ping` berarti host tidak merespon —
		// itu hasil valid (offline), bukan kegagalan pengecekan itu sendiri.
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return Result{Online: false}, nil
		}
		return Result{}, fmt.Errorf("gagal menjalankan ping: %w", err)
	}
	return Result{Online: true, RTT: elapsed}, nil
}
