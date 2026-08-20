// Package selfupdate berisi logika "Cek Update"/"Update Sekarang" di
// Settings > Sistem: mengecek rilis terbaru di GitHub, membandingkan
// dengan versi yang sedang berjalan, dan (kalau diaktifkan lewat
// SelfUpdateConfig.Enabled) menjalankan skrip deploy di latar belakang.
//
// Paket ini SENGAJA tidak bergantung ke internal/ (pola yang sama dengan
// pkg/reportexport) — cuma butuh string versi & konfigurasi, tidak perlu
// tahu apa pun soal model/repository aplikasi.
package selfupdate

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ErrNoReleases dikembalikan kalau repo GitHub belum pernah punya rilis
// resmi (endpoint /releases/latest mengembalikan 404) — beda dari error
// jaringan/permintaan gagal, supaya pemanggil bisa kasih pesan yang tepat
// ("repo ini belum punya rilis") alih-alih pesan error generik.
var ErrNoReleases = errors.New("selfupdate: repo belum punya rilis resmi di GitHub")

// Release — subset field dari response GitHub REST API
// GET /repos/{owner}/{repo}/releases/latest yang relevan di sini.
type Release struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	HTMLURL     string `json:"html_url"`
	Body        string `json:"body"`
	PublishedAt string `json:"published_at"`
	Prerelease  bool   `json:"prerelease"`
	Draft       bool   `json:"draft"`
}

// FetchLatestRelease memanggil GitHub REST API PUBLIK (tidak butuh token
// — cukup untuk repo publik; untuk repo privat, tambahkan header
// Authorization di sini kalau nanti dibutuhkan) untuk mengambil rilis
// TERBARU (non-draft, non-prerelease — itu definisi resmi endpoint
// /releases/latest dari GitHub sendiri, jadi kita tidak perlu menyortir
// manual daftar tag).
func FetchLatestRelease(owner, repo string) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	client := &http.Client{Timeout: 8 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "gowms-selfupdate")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal menghubungi GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNoReleases
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("GitHub API mengembalikan status %d: %s", resp.StatusCode, string(body))
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("gagal membaca response GitHub: %w", err)
	}
	return &release, nil
}

// CompareVersions membandingkan dua string versi ala semver TAPI toleran
// terhadap prefix "v" dan jumlah segmen yang tidak sama (mis. "v1.3" vs
// "v1.3.0" dianggap setara) — tag di GitHub sering ditulis manual, tidak
// selalu 3 segmen persis. Mengembalikan:
//
//	-1 kalau a < b (a lebih lama)
//	 0 kalau a == b
//	 1 kalau a > b (a lebih baru)
//
// Segmen non-angka (mis. "v1.3.0-beta") dipotong di karakter non-angka
// pertama tiap segmen — cukup untuk kebutuhan "ada versi lebih baru atau
// tidak", bukan pembanding semver lengkap dengan pre-release/build
// metadata.
func CompareVersions(a, b string) int {
	segA := normalizeVersionSegments(a)
	segB := normalizeVersionSegments(b)
	maxLen := len(segA)
	if len(segB) > maxLen {
		maxLen = len(segB)
	}
	for i := 0; i < maxLen; i++ {
		va, vb := 0, 0
		if i < len(segA) {
			va = segA[i]
		}
		if i < len(segB) {
			vb = segB[i]
		}
		if va != vb {
			if va < vb {
				return -1
			}
			return 1
		}
	}
	return 0
}

func normalizeVersionSegments(v string) []int {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")
	parts := strings.Split(v, ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		numEnd := 0
		for numEnd < len(p) && p[numEnd] >= '0' && p[numEnd] <= '9' {
			numEnd++
		}
		if numEnd == 0 {
			out = append(out, 0)
			continue
		}
		n, err := strconv.Atoi(p[:numEnd])
		if err != nil {
			n = 0
		}
		out = append(out, n)
	}
	return out
}
