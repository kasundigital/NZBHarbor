package server

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kasundigital/NZBHarbor/internal/config"
	"github.com/kasundigital/NZBHarbor/internal/downloader"
	"github.com/kasundigital/NZBHarbor/internal/model"
	"github.com/kasundigital/NZBHarbor/internal/nntp"
	"github.com/kasundigital/NZBHarbor/internal/store"
)

type Server struct {
	cfg    *config.Config
	store  *store.Store
	engine *downloader.Engine
	mux    *http.ServeMux
}

func New(c *config.Config, s *store.Store, e *downloader.Engine) *Server {
	x := &Server{cfg: c, store: s, engine: e, mux: http.NewServeMux()}
	x.routes()
	return x
}
func (s *Server) Run(ctx context.Context) error {
	h := logMiddleware(s.mux)
	srv := &http.Server{Addr: s.cfg.Listen, Handler: h, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		<-ctx.Done()
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(c)
	}()
	log.Printf("NZBHarbor listening on %s", s.cfg.Listen)
	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}
func (s *Server) routes() {
	s.mux.HandleFunc("/api", s.sab)
	s.mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		jsonOut(w, map[string]any{"ok": true, "name": "NZBHarbor", "version": "0.1.0"})
	})
	s.mux.HandleFunc("/api/v1/status", s.auth(s.status))
	s.mux.HandleFunc("/api/v1/jobs", s.auth(s.jobs))
	s.mux.HandleFunc("/api/v1/jobs/", s.auth(s.jobAction))
	s.mux.HandleFunc("/api/v1/config", s.auth(s.configAPI))
	s.mux.HandleFunc("/api/v1/test-server", s.auth(s.testServer))
	fs := http.FileServer(http.Dir(s.cfg.WebDir))
	s.mux.Handle("/", fs)
}
func (s *Server) status(w http.ResponseWriter, r *http.Request) {
	jobs := s.store.List()
	var q, h int
	for _, j := range jobs {
		if j.Status == "completed" || j.Status == "failed" {
			h++
		} else {
			q++
		}
	}
	jsonOut(w, map[string]any{"queue": q, "history": h, "servers": len(s.cfg.Servers), "download_dir": s.cfg.DownloadDir})
}
func (s *Server) jobs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		jsonOut(w, s.store.List())
	case http.MethodPost:
		s.upload(w, r)
	default:
		http.Error(w, "method", 405)
	}
}
func (s *Server) upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(128 << 20); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	f, h, err := firstFile(r.MultipartForm)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	defer f.Close()
	name := r.FormValue("name")
	if name == "" {
		name = strings.TrimSuffix(h.Filename, filepath.Ext(h.Filename))
	}
	j, err := s.engine.Add(name, r.FormValue("category"), f)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(201)
	jsonOut(w, j)
}
func (s *Server) jobAction(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimPrefix(r.URL.Path, "/api/v1/jobs/")
	x := strings.Split(strings.Trim(p, "/"), "/")
	if len(x) < 2 {
		http.Error(w, "action required", 400)
		return
	}
	var err error
	switch x[1] {
	case "pause":
		err = s.engine.Pause(x[0])
	case "resume":
		err = s.engine.Resume(x[0])
	case "delete":
		err = s.engine.Delete(x[0])
	default:
		http.Error(w, "unknown action", 404)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	jsonOut(w, map[string]bool{"ok": true})
}
func (s *Server) configAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		cp := *s.cfg
		for i := range cp.Servers {
			if cp.Servers[i].Password != "" {
				cp.Servers[i].Password = "********"
			}
		}
		jsonOut(w, cp)
		return
	}
	if r.Method != http.MethodPut {
		http.Error(w, "method", 405)
		return
	}
	var in config.Config
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	for i := range in.Servers {
		if in.Servers[i].Password == "********" && i < len(s.cfg.Servers) {
			in.Servers[i].Password = s.cfg.Servers[i].Password
		}
	}
	in.ConfigDir = s.cfg.ConfigDir
	in.DownloadDir = s.cfg.DownloadDir
	in.WebDir = s.cfg.WebDir
	if in.Listen == "" {
		in.Listen = s.cfg.Listen
	}
	if in.APIKey == "" {
		in.APIKey = s.cfg.APIKey
	}
	if in.TempDir == "" {
		in.TempDir = s.cfg.TempDir
	}
	*s.cfg = in
	if err := config.Save(s.cfg); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	jsonOut(w, map[string]bool{"ok": true})
}
func (s *Server) testServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	var srv config.NewsServer
	if err := json.NewDecoder(r.Body).Decode(&srv); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if srv.Password == "********" {
		for _, existing := range s.cfg.Servers {
			if existing.Host == srv.Host && existing.Username == srv.Username {
				srv.Password = existing.Password
				break
			}
		}
	}
	client, err := nntp.Dial(srv)
	if err != nil {
		jsonOut(w, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	_ = client.Close()
	jsonOut(w, map[string]any{"ok": true, "message": "NNTP connection and authentication succeeded"})
}

func (s *Server) sab(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseMultipartForm(128 << 20)
	if !s.sabAuth(r) {
		jsonOut(w, map[string]any{"status": false, "error": "API Key Incorrect"})
		return
	}
	mode := r.FormValue("mode")
	switch mode {
	case "version":
		jsonOut(w, map[string]string{"version": "5.1.0"})
	case "get_config":
		cats := []map[string]any{}
		for _, name := range []string{"default", "tv", "movies", "music"} {
			cats = append(cats, map[string]any{"name": name, "priority": 0, "pp": "", "script": "None", "dir": name})
		}
		jsonOut(w, map[string]any{"config": map[string]any{
			"misc":       map[string]any{"complete_dir": filepath.Join(s.cfg.DownloadDir, "complete"), "pre_check": false, "history_retention": "", "history_retention_option": "all", "history_retention_number": 0},
			"categories": cats, "servers": []any{}, "sorters": []any{},
		}})
	case "fullstatus":
		jsonOut(w, map[string]any{"status": map[string]any{"completedir": filepath.Join(s.cfg.DownloadDir, "complete")}})
	case "queue":
		s.sabQueue(w, r)
	case "history":
		s.sabHistory(w, r)
	case "retry":
		id := r.FormValue("value")
		err := s.engine.Resume(id)
		jsonOut(w, map[string]any{"status": err == nil, "nzo_ids": []string{id}})
	case "addfile":
		f, h, err := firstFile(r.MultipartForm)
		if err != nil {
			jsonOut(w, map[string]any{"status": false, "error": err.Error()})
			return
		}
		defer f.Close()
		name := r.FormValue("nzbname")
		if name == "" {
			name = strings.TrimSuffix(h.Filename, filepath.Ext(h.Filename))
		}
		j, err := s.engine.Add(name, r.FormValue("cat"), f)
		if err != nil {
			jsonOut(w, map[string]any{"status": false, "error": err.Error()})
			return
		}
		jsonOut(w, map[string]any{"status": true, "nzo_ids": []string{j.ID}})
	default:
		jsonOut(w, map[string]any{"status": false, "error": "Unsupported SAB mode: " + mode})
	}
}
func (s *Server) sabQueue(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("name") != "" {
		s.sabQueueAction(w, r)
		return
	}
	slots := []map[string]any{}
	category := r.FormValue("category")
	for _, j := range s.store.List() {
		if j.Status == "completed" || j.Status == "failed" {
			continue
		}
		if category != "" && j.Category != category {
			continue
		}
		slots = append(slots, sabSlot(j))
	}
	jsonOut(w, map[string]any{"queue": map[string]any{"status": "Downloading", "paused": false, "speed": "0", "kbpersec": "0", "sizeleft": "0 B", "mb": "0", "mbleft": "0", "noofslots": len(slots), "slots": slots}})
}
func (s *Server) sabHistory(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("name") == "delete" {
		id := r.FormValue("value")
		if j, ok := s.store.Get(id); ok && r.FormValue("del_files") == "1" && j.Storage != "" {
			_ = os.RemoveAll(j.Storage)
		}
		err := s.engine.Delete(id)
		jsonOut(w, map[string]any{"status": err == nil})
		return
	}
	slots := []map[string]any{}
	category := r.FormValue("category")
	for _, j := range s.store.List() {
		if j.Status != "completed" && j.Status != "failed" {
			continue
		}
		if category != "" && j.Category != category {
			continue
		}
		completed := j.CompletedAt
		if completed.IsZero() {
			completed = j.UpdatedAt
		}
		downloadTime := int(j.UpdatedAt.Sub(j.CreatedAt).Seconds())
		if downloadTime < 0 {
			downloadTime = 0
		}
		slots = append(slots, map[string]any{"nzo_id": j.ID, "name": j.Name, "nzb_name": j.Name, "bytes": j.TotalBytes, "download_time": downloadTime, "status": strings.Title(j.Status), "category": j.Category, "storage": j.Storage, "fail_message": j.Error, "completed": completed.Unix()})
	}
	jsonOut(w, map[string]any{"history": map[string]any{"noofslots": len(slots), "slots": slots}})
}
func (s *Server) sabQueueAction(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	id := r.FormValue("value")
	var err error
	switch name {
	case "pause":
		if id != "" {
			err = s.engine.Pause(id)
		}
	case "resume":
		if id != "" {
			err = s.engine.Resume(id)
		}
	case "delete":
		if id != "" {
			err = s.engine.Delete(id)
		}
	default:
		err = fmt.Errorf("unsupported action")
	}
	jsonOut(w, map[string]any{"status": err == nil})
}
func sabSlot(j model.Job) map[string]any {
	return map[string]any{"nzo_id": j.ID, "filename": j.Name, "status": strings.Title(j.Status), "percentage": fmt.Sprintf("%.1f", j.Progress), "mb": fmt.Sprintf("%.1f", float64(j.TotalBytes)/(1024*1024)), "mbleft": fmt.Sprintf("%.1f", float64(j.TotalBytes-j.DoneBytes)/(1024*1024)), "cat": j.Category, "size": fmt.Sprintf("%d", j.TotalBytes), "sizeleft": fmt.Sprintf("%d", j.TotalBytes-j.DoneBytes)}
}
func (s *Server) sabAuth(r *http.Request) bool {
	if s.cfg.APIKey == "" {
		return true
	}
	k := r.FormValue("apikey")
	return subtle.ConstantTimeCompare([]byte(k), []byte(s.cfg.APIKey)) == 1
}
func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.APIKey == "" {
			next(w, r)
			return
		}
		k := r.Header.Get("X-Api-Key")
		if k == "" {
			k = r.URL.Query().Get("apikey")
		}
		if subtle.ConstantTimeCompare([]byte(k), []byte(s.cfg.APIKey)) != 1 {
			http.Error(w, "unauthorized", 401)
			return
		}
		next(w, r)
	}
}
func firstFile(m *multipart.Form) (multipart.File, *multipart.FileHeader, error) {
	if m == nil {
		return nil, nil, fmt.Errorf("multipart upload required")
	}
	for _, hs := range m.File {
		if len(hs) > 0 {
			f, e := hs[0].Open()
			return f, hs[0], e
		}
	}
	return nil, nil, fmt.Errorf("NZB file missing")
}
func jsonOut(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
func logMiddleware(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		n.ServeHTTP(w, r)
	})
}

var _ = io.Copy
var _ = os.ErrNotExist
var _ = strconv.Itoa
