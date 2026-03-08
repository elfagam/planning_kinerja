package repository

import (
	"sync"
	"time"

	"e-plan-ai/internal/modules/crud/domain"
)

type resourceStore struct {
	nextID int64
	items  map[int64]domain.Record
}

type MemoryStore struct {
	mu        sync.RWMutex
	resources map[string]*resourceStore
}

func NewMemoryStore(resourceKeys []string) *MemoryStore {
	resources := map[string]*resourceStore{}
	for _, key := range resourceKeys {
		resources[key] = &resourceStore{nextID: 1, items: map[int64]domain.Record{}}
	}
	return &MemoryStore{resources: resources}
}

func (s *MemoryStore) List(resource string, filter domain.ListFilter) ([]domain.Record, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	store := s.resources[resource]

	rows := make([]domain.Record, 0, len(store.items))
	for _, item := range store.items {
		if filter.Query != "" && !containsInsensitive(item.Name, filter.Query) && !containsInsensitive(item.Code, filter.Query) {
			continue
		}
		rows = append(rows, item)
	}

	total := int64(len(rows))
	if filter.Offset >= len(rows) {
		return []domain.Record{}, total, nil
	}
	end := filter.Offset + filter.Limit
	if end > len(rows) {
		end = len(rows)
	}
	return rows[filter.Offset:end], total, nil
}

func (s *MemoryStore) Create(resource string, payload domain.Payload) (domain.Record, error) {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()

	store := s.resources[resource]
	id := store.nextID
	store.nextID++

	item := domain.Record{
		ID:          id,
		Code:        payload.Code,
		Name:        payload.Name,
		Description: payload.Description,
		Attributes:  payload.Attributes,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	store.items[id] = item
	return item, nil
}

func (s *MemoryStore) Get(resource string, id int64) (domain.Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.resources[resource].items[id]
	if !ok {
		return domain.Record{}, ErrNotFound
	}
	return item, nil
}

func (s *MemoryStore) Update(resource string, id int64, payload domain.Payload) (domain.Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	store := s.resources[resource]
	item, ok := store.items[id]
	if !ok {
		return domain.Record{}, ErrNotFound
	}

	item.Code = payload.Code
	item.Name = payload.Name
	item.Description = payload.Description
	item.Attributes = payload.Attributes
	item.UpdatedAt = time.Now()
	store.items[id] = item
	return item, nil
}

func (s *MemoryStore) Delete(resource string, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	store := s.resources[resource]
	if _, ok := store.items[id]; !ok {
		return ErrNotFound
	}
	delete(store.items, id)
	return nil
}

func containsInsensitive(source, query string) bool {
	if query == "" {
		return true
	}
	s := []rune(source)
	q := []rune(query)
	for i := range s {
		if i+len(q) > len(s) {
			break
		}
		match := true
		for j := range q {
			a := s[i+j]
			b := q[j]
			if 'A' <= a && a <= 'Z' {
				a = a + 32
			}
			if 'A' <= b && b <= 'Z' {
				b = b + 32
			}
			if a != b {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
