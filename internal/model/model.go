package model

import "time"

type Segment struct {
	Number int    `xml:"number,attr" json:"number"`
	Bytes  int64  `xml:"bytes,attr" json:"bytes"`
	ID     string `xml:",chardata" json:"id"`
}

type NZBFile struct {
	Subject  string    `xml:"subject,attr" json:"subject"`
	Date     int64     `xml:"date,attr" json:"date"`
	Filename string    `json:"filename"`
	Groups   []string  `xml:"groups>group" json:"groups"`
	Segments []Segment `xml:"segments>segment" json:"segments"`
}

type NZB struct {
	Files []NZBFile `xml:"file"`
}

type Job struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Status      string    `json:"status"`
	Progress    float64   `json:"progress"`
	TotalBytes  int64     `json:"total_bytes"`
	DoneBytes   int64     `json:"done_bytes"`
	Speed       int64     `json:"speed"`
	Error       string    `json:"error,omitempty"`
	Storage     string    `json:"storage,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	NZBPath     string    `json:"nzb_path"`
}
