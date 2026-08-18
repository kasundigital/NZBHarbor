package nzb

import (
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/kasundigital/NZBHarbor/internal/model"
)

var quoted = regexp.MustCompile(`"([^"]+)"`)
var unsafe = regexp.MustCompile(`[^a-zA-Z0-9._()\[\] -]+`)

func Parse(r io.Reader) (*model.NZB, error) {
	var doc model.NZB
	if err := xml.NewDecoder(r).Decode(&doc); err != nil {
		return nil, fmt.Errorf("parse nzb: %w", err)
	}
	if len(doc.Files) == 0 {
		return nil, fmt.Errorf("nzb contains no files")
	}
	for i := range doc.Files {
		doc.Files[i].Filename = filename(doc.Files[i].Subject, i)
		sort.Slice(doc.Files[i].Segments, func(a, b int) bool { return doc.Files[i].Segments[a].Number < doc.Files[i].Segments[b].Number })
		for j := range doc.Files[i].Segments {
			doc.Files[i].Segments[j].ID = strings.TrimSpace(doc.Files[i].Segments[j].ID)
		}
	}
	return &doc, nil
}

func filename(subject string, i int) string {
	if m := quoted.FindStringSubmatch(subject); len(m) == 2 {
		return clean(m[1])
	}
	p := strings.Fields(subject)
	if len(p) > 0 {
		return clean(p[0])
	}
	return fmt.Sprintf("file-%03d.bin", i+1)
}

func clean(s string) string {
	s = unsafe.ReplaceAllString(strings.TrimSpace(s), "_")
	s = strings.Trim(s, ". ")
	if s == "" {
		return "download.bin"
	}
	return s
}
