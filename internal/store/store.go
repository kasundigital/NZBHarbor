package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/kasundigital/NZBHarbor/internal/model"
)

type Store struct {
	mu   sync.RWMutex
	path string
	jobs map[string]*model.Job
}

func New(configDir string) (*Store, error) {
	s := &Store{path: filepath.Join(configDir, "state.json"), jobs: map[string]*model.Job{}}
	b, err := os.ReadFile(s.path)
	if err == nil {
		if err := json.Unmarshal(b, &s.jobs); err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	// Interrupted work is safe to resume because segment files are persisted.
	recovered := false
	for _, j := range s.jobs {
		switch j.Status {
		case "downloading", "post-processing":
			j.Status = "queued"
			j.Error = ""
			j.Speed = 0
			j.UpdatedAt = time.Now()
			recovered = true
		}
	}
	if recovered {
		if err := s.flushUnlocked(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *Store) Save(j *model.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *j
	s.jobs[j.ID] = &cp
	return s.flushUnlocked()
}

func (s *Store) Get(id string) (*model.Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, ok := s.jobs[id]
	if !ok {
		return nil, false
	}
	cp := *j
	return &cp, true
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.jobs, id)
	return s.flushUnlocked()
}

func (s *Store) List() []model.Job {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		out = append(out, *j)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *Store) flushUnlocked() error {
	b, err := json.MarshalIndent(s.jobs, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if _, err := f.Write(b); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, s.path)
}
