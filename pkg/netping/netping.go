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

type Result struct {
	Online bool

	RTT time.Duration
}

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

	cmd := exec.CommandContext(ctx, "ping", "-c", "1", "-W", waitSecs, target)
	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)

	if ctx.Err() != nil {
		return Result{}, fmt.Errorf("ping ke %s timeout: %w", target, ctx.Err())
	}
	if err != nil {

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return Result{Online: false}, nil
		}
		return Result{}, fmt.Errorf("gagal menjalankan ping: %w", err)
	}
	return Result{Online: true, RTT: elapsed}, nil
}
