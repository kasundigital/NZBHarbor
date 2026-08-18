package nntp

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func DecodeYEnc(article []byte) ([]byte, error) {
	lines := bytes.Split(article, []byte("\n"))
	started := false
	var out []byte
	for _, raw := range lines {
		line := strings.TrimSuffix(string(raw), "\r")
		if strings.HasPrefix(line, "=ybegin") {
			started = true
			continue
		}
		if !started {
			continue
		}
		if strings.HasPrefix(line, "=ypart") {
			continue
		}
		if strings.HasPrefix(line, "=yend") {
			return out, nil
		}
		b := []byte(line)
		for i := 0; i < len(b); i++ {
			v := b[i]
			if v == '=' {
				i++
				if i >= len(b) {
					return nil, fmt.Errorf("invalid yenc escape")
				}
				v = b[i] - 64
			}
			out = append(out, v-42)
		}
	}
	if !started {
		return nil, fmt.Errorf("article is not yEnc encoded")
	}
	return nil, fmt.Errorf("missing yEnc end marker; decoded %s bytes", strconv.Itoa(len(out)))
}
