package nntp

import (
	"bytes"
	"fmt"
	"hash/crc32"
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
			fields := parseYEncFields(line)
			if sizeText := fields["size"]; sizeText != "" {
				size, err := strconv.Atoi(sizeText)
				if err != nil {
					return nil, fmt.Errorf("invalid yEnc size %q", sizeText)
				}
				if size != len(out) {
					return nil, fmt.Errorf("yEnc size mismatch: expected %d, got %d", size, len(out))
				}
			}
			expectedCRC := fields["pcrc32"]
			if expectedCRC == "" {
				expectedCRC = fields["crc32"]
			}
			if expectedCRC != "" {
				want, err := strconv.ParseUint(expectedCRC, 16, 32)
				if err != nil {
					return nil, fmt.Errorf("invalid yEnc CRC %q", expectedCRC)
				}
				got := crc32.ChecksumIEEE(out)
				if got != uint32(want) {
					return nil, fmt.Errorf("yEnc CRC mismatch: expected %08x, got %08x", uint32(want), got)
				}
			}
			return out, nil
		}

		b := []byte(line)
		for i := 0; i < len(b); i++ {
			v := b[i]
			if v == '=' {
				i++
				if i >= len(b) {
					return nil, fmt.Errorf("invalid yEnc escape")
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

func parseYEncFields(line string) map[string]string {
	out := map[string]string{}
	for _, field := range strings.Fields(line) {
		if key, value, ok := strings.Cut(field, "="); ok {
			out[strings.ToLower(key)] = strings.TrimSpace(value)
		}
	}
	return out
}
