package postprocess

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const commandTimeout = 2 * time.Hour

func Run(ctx context.Context, dir string) error {
	pars, _ := filepath.Glob(filepath.Join(dir, "*.par2"))
	sort.Strings(pars)
	if len(pars) > 0 {
		mainPAR := pars[0]
		for _, p := range pars {
			if !strings.Contains(strings.ToLower(filepath.Base(p)), ".vol") {
				mainPAR = p
				break
			}
		}
		p, err := exec.LookPath("par2")
		if err != nil {
			return fmt.Errorf("par2 executable not found: %w", err)
		}
		runCtx, cancel := context.WithTimeout(ctx, commandTimeout)
		out, err := exec.CommandContext(runCtx, p, "r", mainPAR).CombinedOutput()
		cancel()
		if err != nil {
			if runCtx.Err() != nil {
				return fmt.Errorf("par2 repair timed out or was cancelled: %w", runCtx.Err())
			}
			return fmt.Errorf("par2 repair: %v: %s", err, strings.TrimSpace(string(out)))
		}
	}

	entries, _ := filepath.Glob(filepath.Join(dir, "*"))
	sort.Strings(entries)
	archive := ""
	for _, p := range entries {
		name := strings.ToLower(filepath.Base(p))
		if strings.HasSuffix(name, ".part01.rar") {
			archive = p
			break
		}
		if strings.HasSuffix(name, ".rar") && !strings.Contains(name, ".part") && archive == "" {
			archive = p
		}
	}
	if archive == "" {
		return nil
	}

	u, err := exec.LookPath("unar")
	if err != nil {
		return fmt.Errorf("unar executable not found: %w", err)
	}
	runCtx, cancel := context.WithTimeout(ctx, commandTimeout)
	out, err := exec.CommandContext(runCtx, u, "-f", "-o", dir, archive).CombinedOutput()
	cancel()
	if err != nil {
		if runCtx.Err() != nil {
			return fmt.Errorf("archive extraction timed out or was cancelled: %w", runCtx.Err())
		}
		return fmt.Errorf("extract: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
