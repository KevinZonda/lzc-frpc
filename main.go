package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type todoItem struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Done      bool   `json:"done"`
	UpdatedAt int64  `json:"updatedAt"`
}

type todoStore struct {
	mu     sync.RWMutex
	nextID atomic.Int64
	items  map[int64]todoItem
	file   string
}

func newTodoStore(file string) (*todoStore, error) {
	store := &todoStore{
		items: map[int64]todoItem{},
		file:  file,
	}
	store.nextID.Store(1)
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return nil, err
	}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *todoStore) load() error {
	bs, err := os.ReadFile(s.file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var items []todoItem
	if len(bs) == 0 {
		return nil
	}
	if err := json.Unmarshal(bs, &items); err != nil {
		return err
	}

	var maxID int64 = 0
	for _, item := range items {
		s.items[item.ID] = item
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	s.nextID.Store(maxID + 1)
	return nil
}

func (s *todoStore) persistLocked() error {
	items := make([]todoItem, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})

	bs, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	tmpFile := s.file + ".tmp"
	if err := os.WriteFile(tmpFile, bs, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpFile, s.file)
}

func (s *todoStore) list() []todoItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]todoItem, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].UpdatedAt > items[j].UpdatedAt
	})
	return items
}

func (s *todoStore) add(title string) (todoItem, error) {
	id := s.nextID.Add(1) - 1
	item := todoItem{
		ID:        id,
		Title:     title,
		Done:      false,
		UpdatedAt: time.Now().UnixMilli(),
	}
	s.mu.Lock()
	s.items[id] = item
	err := s.persistLocked()
	s.mu.Unlock()
	if err != nil {
		return todoItem{}, err
	}
	return item, nil
}

func (s *todoStore) toggle(id int64) (todoItem, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	if !ok {
		return todoItem{}, false, nil
	}
	item.Done = !item.Done
	item.UpdatedAt = time.Now().UnixMilli()
	s.items[id] = item
	if err := s.persistLocked(); err != nil {
		return todoItem{}, false, err
	}
	return item, true, nil
}

func (s *todoStore) remove(id int64) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return false, nil
	}
	delete(s.items, id)
	if err := s.persistLocked(); err != nil {
		return false, err
	}
	return true, nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func main() {
	store, err := newTodoStore("/lzcapp/var/todos.json")
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/api/todos", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, map[string]any{"items": store.list()})
		case http.MethodPost:
			var payload struct {
				Title string `json:"title"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
				return
			}
			title := strings.TrimSpace(payload.Title)
			if title == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
				return
			}
			item, err := store.add(title)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to persist todos"})
				return
			}
			writeJSON(w, http.StatusOK, item)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/todos/", func(w http.ResponseWriter, r *http.Request) {
		tail := strings.TrimPrefix(r.URL.Path, "/api/todos/")
		if strings.HasSuffix(tail, "/toggle") {
			if r.Method != http.MethodPut {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			idPart := strings.TrimSuffix(tail, "/toggle")
			idPart = strings.TrimSuffix(idPart, "/")
			id, err := strconv.ParseInt(idPart, 10, 64)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
				return
			}
			item, ok, err := store.toggle(id)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to persist todos"})
				return
			}
			if !ok {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "todo not found"})
				return
			}
			writeJSON(w, http.StatusOK, item)
			return
		}

		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		idPart := strings.TrimSuffix(tail, "/")
		id, err := strconv.ParseInt(idPart, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
			return
		}
		removed, err := store.remove(id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to persist todos"})
			return
		}
		if !removed {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "todo not found"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})

	mux.Handle("/", http.FileServer(http.Dir("/app/web")))
	_ = http.ListenAndServe(":3000", mux)
}
