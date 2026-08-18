package postprocess

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func Run(dir string) error {
	pars, _ := filepath.Glob(filepath.Join(dir, "*.par2"))
	if len(pars) > 0 {
		if p, err := exec.LookPath("par2"); err == nil {
			if out, err := exec.Command(p, "r", pars[0]).CombinedOutput(); err != nil {
				return fmt.Errorf("par2 repair: %v: %s", err, out)
			}
		}
	}
	entries, _ := filepath.Glob(filepath.Join(dir, "*"))
	for _, p := range entries {
		l := strings.ToLower(p)
		if strings.HasSuffix(l, ".rar") && !strings.Contains(l, ".part") || strings.HasSuffix(l, ".part01.rar") {
			if u, err := exec.LookPath("unar"); err == nil {
				if out, err := exec.Command(u, "-f", "-o", dir, p).CombinedOutput(); err != nil {
					return fmt.Errorf("extract: %v: %s", err, out)
				}
				break
			}
		}
	}
	return nil
}
