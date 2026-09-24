package downloader

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kasundigital/NZBHarbor/internal/config"
	"github.com/kasundigital/NZBHarbor/internal/model"
	"github.com/kasundigital/NZBHarbor/internal/nntp"
	nzbparser "github.com/kasundigital/NZBHarbor/internal/nzb"
	"github.com/kasundigital/NZBHarbor/internal/postprocess"
	"github.com/kasundigital/NZBHarbor/internal/store"
)

const articleAttemptsPerProvider = 3

type Engine struct {
	cfg    *config.Config
	store  *store.Store
	wake   chan struct{}
	mu     sync.Mutex
	cancel map[string]context.CancelFunc

	poolMu sync.Mutex
	pools  map[string]*nntp.Pool
}

func New(cfg *config.Config, st *store.Store) *Engine {
	return &Engine{
		cfg: cfg, store: st, wake: make(chan struct{}, 1),
		cancel: map[string]context.CancelFunc{},
		pools:  map[string]*nntp.Pool{},
	}
}

func (e *Engine) Run(ctx context.Context) {
	defer e.closePools()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.tryStart(ctx)
		case <-e.wake:
			e.tryStart(ctx)
		}
	}
}

func (e *Engine) Signal() {
	select {
	case e.wake <- struct{}{}:
	default:
	}
}

func (e *Engine) Add(name, category string, r io.Reader) (*model.Job, error) {
	if category == "" {
		category = "default"
	}
	jobID := newID()
	if name == "" {
		name = "NZB-" + jobID[:8]
	}

	dir := filepath.Join(e.cfg.ConfigDir, "nzbs")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, jobID+".nzb")
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(f, r); err != nil {
		_ = f.Close()
		return nil, err
	}
	_ = f.Close()

	pf, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	doc, err := nzbparser.Parse(pf)
	_ = pf.Close()
	if err != nil {
		_ = os.Remove(path)
		return nil, err
	}

	var total int64
	for _, file := range doc.Files {
		for _, seg := range file.Segments {
			total += seg.Bytes
		}
	}
	now := time.Now()
	j := &model.Job{
		ID: jobID, Name: safe(name), Category: safe(category), Status: "queued",
		TotalBytes: total, CreatedAt: now, UpdatedAt: now, NZBPath: path,
	}
	if err := e.store.Save(j); err != nil {
		return nil, err
	}
	e.Signal()
	return j, nil
}

func (e *Engine) Pause(id string) error {
	j, ok := e.store.Get(id)
	if !ok {
		return fmt.Errorf("job not found")
	}
	e.mu.Lock()
	if cancel := e.cancel[id]; cancel != nil {
		cancel()
	}
	e.mu.Unlock()
	j.Status = "paused"
	j.Speed = 0
	j.UpdatedAt = time.Now()
	return e.store.Save(j)
}

func (e *Engine) Resume(id string) error {
	j, ok := e.store.Get(id)
	if !ok {
		return fmt.Errorf("job not found")
	}
	j.Status = "queued"
	j.Error = ""
	j.Speed = 0
	j.UpdatedAt = time.Now()
	if err := e.store.Save(j); err != nil {
		return err
	}
	e.Signal()
	return nil
}

func (e *Engine) Delete(id string) error {
	e.mu.Lock()
	if cancel := e.cancel[id]; cancel != nil {
		cancel()
	}
	delete(e.cancel, id)
	e.mu.Unlock()
	return e.store.Delete(id)
}

func (e *Engine) tryStart(parent context.Context) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.cancel) > 0 {
		return
	}
	for _, j := range e.store.List() {
		if j.Status == "queued" {
			ctx, cancel := context.WithCancel(parent)
			e.cancel[j.ID] = cancel
			go e.process(ctx, j.ID)
			return
		}
	}
}

func (e *Engine) process(ctx context.Context, id string) {
	defer func() {
		e.mu.Lock()
		delete(e.cancel, id)
		e.mu.Unlock()
		e.Signal()
	}()

	j, ok := e.store.Get(id)
	if !ok {
		return
	}
	j.Status = "downloading"
	j.UpdatedAt = time.Now()
	_ = e.store.Save(j)

	if err := e.download(ctx, j); err != nil {
		if ctx.Err() != nil {
			return
		}
		j.Status = "failed"
		j.Speed = 0
		j.Error = err.Error()
		j.UpdatedAt = time.Now()
		_ = e.store.Save(j)
		return
	}

	if e.cfg.PostProcess {
		j.Status = "post-processing"
		j.Speed = 0
		j.UpdatedAt = time.Now()
		_ = e.store.Save(j)
		if err := postprocess.Run(ctx, j.Storage); err != nil {
			if ctx.Err() != nil {
				return
			}
			j.Status = "failed"
			j.Error = err.Error()
			j.UpdatedAt = time.Now()
			_ = e.store.Save(j)
			return
		}
	}

	j.Status = "completed"
	j.Progress = 100
	j.DoneBytes = j.TotalBytes
	j.Speed = 0
	j.CompletedAt = time.Now()
	j.UpdatedAt = time.Now()
	_ = e.store.Save(j)
}

func (e *Engine) download(ctx context.Context, j *model.Job) error {
	if len(e.cfg.Servers) == 0 {
		return fmt.Errorf("no Usenet servers configured")
	}
	f, err := os.Open(j.NZBPath)
	if err != nil {
		return err
	}
	doc, err := nzbparser.Parse(f)
	_ = f.Close()
	if err != nil {
		return err
	}

	outDir := filepath.Join(e.cfg.DownloadDir, "complete", j.Category, j.Name)
	tmpDir := filepath.Join(e.cfg.TempDir, j.ID)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return err
	}
	j.Storage = outDir

	var done atomic.Int64
	started := time.Now()
	var progressMu sync.Mutex

	for fileIndex, file := range doc.Files {
		if err := ctx.Err(); err != nil {
			return err
		}
		partsDir := filepath.Join(tmpDir, fmt.Sprintf("%04d", fileIndex))
		if err := os.MkdirAll(partsDir, 0755); err != nil {
			return err
		}

		type task struct {
			seg  model.Segment
			path string
		}
		tasks := make(chan task, len(file.Segments))
		errs := make(chan error, 1)
		workers := e.cfg.MaxWorkers
		if workers > len(file.Segments) {
			workers = len(file.Segments)
		}
		if workers < 1 {
			workers = 1
		}
		var wg sync.WaitGroup
		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for t := range tasks {
					if ctx.Err() != nil {
						return
					}
					if st, statErr := os.Stat(t.path); statErr == nil && st.Size() > 0 {
						e.updateProgress(j, &done, t.seg.Bytes, started, &progressMu)
						continue
					}
					data, fetchErr := e.fetch(ctx, t.seg.ID)
					if fetchErr != nil {
						select {
						case errs <- fetchErr:
						default:
						}
						return
					}
					if writeErr := os.WriteFile(t.path, data, 0644); writeErr != nil {
						select {
						case errs <- writeErr:
						default:
						}
						return
					}
					e.updateProgress(j, &done, t.seg.Bytes, started, &progressMu)
				}
			}()
		}

		for _, seg := range file.Segments {
			tasks <- task{seg: seg, path: filepath.Join(partsDir, fmt.Sprintf("%08d.part", seg.Number))}
		}
		close(tasks)
		wg.Wait()
		if err := ctx.Err(); err != nil {
			return err
		}
		select {
		case workerErr := <-errs:
			return workerErr
		default:
		}

		destPath := filepath.Join(outDir, file.Filename)
		dest, err := os.Create(destPath)
		if err != nil {
			return err
		}
		for _, seg := range file.Segments {
			partPath := filepath.Join(partsDir, fmt.Sprintf("%08d.part", seg.Number))
			src, openErr := os.Open(partPath)
			if openErr != nil {
				_ = dest.Close()
				return openErr
			}
			_, copyErr := io.Copy(dest, src)
			_ = src.Close()
			if copyErr != nil {
				_ = dest.Close()
				return copyErr
			}
		}
		if err := dest.Sync(); err != nil {
			_ = dest.Close()
			return err
		}
		if err := dest.Close(); err != nil {
			return err
		}
	}

	if e.cfg.CleanupTemp {
		_ = os.RemoveAll(tmpDir)
	}
	return nil
}

func (e *Engine) updateProgress(j *model.Job, done *atomic.Int64, increment int64, started time.Time, mu *sync.Mutex) {
	current := done.Add(increment)
	mu.Lock()
	defer mu.Unlock()
	j.DoneBytes = current
	if j.TotalBytes > 0 {
		j.Progress = float64(current) * 100 / float64(j.TotalBytes)
	}
	elapsed := time.Since(started).Seconds()
	if elapsed > 0 {
		j.Speed = int64(float64(current) / elapsed)
	}
	j.UpdatedAt = time.Now()
	_ = e.store.Save(j)
}

func (e *Engine) fetch(ctx context.Context, messageID string) ([]byte, error) {
	var last error
	servers := append([]config.NewsServer(nil), e.cfg.Servers...)
	sort.SliceStable(servers, func(i, j int) bool { return servers[i].Priority < servers[j].Priority })

	for _, srv := range servers {
		if !srv.Enabled {
			continue
		}
		pool := e.poolFor(srv)

		for attempt := 1; attempt <= articleAttemptsPerProvider; attempt++ {
			body, err := pool.Body(ctx, messageID)
			if err != nil {
				last = fmt.Errorf("%s article attempt %d/%d: %w", srv.Name, attempt, articleAttemptsPerProvider, err)
				if nntp.IsArticleMissing(err) {
					break
				}
				if attempt < articleAttemptsPerProvider {
					if err := retryDelay(ctx, attempt); err != nil {
						return nil, err
					}
				}
				continue
			}

			data, err := nntp.DecodeYEnc(body)
			if err == nil {
				return data, nil
			}
			last = fmt.Errorf("%s decode attempt %d/%d: %w", srv.Name, attempt, articleAttemptsPerProvider, err)
			if attempt < articleAttemptsPerProvider {
				if err := retryDelay(ctx, attempt); err != nil {
					return nil, err
				}
			}
		}
	}
	if last == nil {
		last = fmt.Errorf("no enabled Usenet server")
	}
	return nil, last
}

func (e *Engine) poolFor(srv config.NewsServer) *nntp.Pool {
	key := fmt.Sprintf("%s|%s|%d|%t|%s|%s|%d", srv.Name, srv.Host, srv.Port, srv.TLS, srv.Username, srv.Password, srv.Connections)
	e.poolMu.Lock()
	defer e.poolMu.Unlock()
	if pool := e.pools[key]; pool != nil {
		return pool
	}
	pool := nntp.NewPool(srv)
	e.pools[key] = pool
	return pool
}

func (e *Engine) closePools() {
	e.poolMu.Lock()
	pools := e.pools
	e.pools = map[string]*nntp.Pool{}
	e.poolMu.Unlock()
	for _, pool := range pools {
		pool.Close()
	}
}

func retryDelay(ctx context.Context, attempt int) error {
	delay := time.Duration(attempt) * 500 * time.Millisecond
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func safe(s string) string {
	s = strings.TrimSpace(s)
	r := strings.NewReplacer("/", "_", "\\", "_", "..", "_")
	s = r.Replace(s)
	if s == "" {
		return "default"
	}
	return s
}
