package byoc_project

import (
	"sync"
)

type Store struct {
	store map[string]*DataplaneResponse
	mu    sync.Mutex
}

func NewStore() *Store {
	return &Store{
		store: make(map[string]*DataplaneResponse),
		mu:    sync.Mutex{},
	}
}

func (s *Store) Get(key string) *DataplaneResponse {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.store[key]
}

// add a dataplane to the store
func (s *Store) Add(dataplane *DataplaneResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.store[dataplane.ProjectID]; !ok {
		s.store[dataplane.ProjectID] = dataplane
	}
}

func (s *Store) Update(key string, dataplane *DataplaneResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[key] = dataplane
}

// update the status of the dataplane
func (s *Store) UpdateStatus(key string, status int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if dataplane, ok := s.store[key]; ok {
		dataplane.Status = status
	}
}

var (
	settingsStore  = NewSafeStore[SettingsResponse]()
	dataplaneStore = NewSafeStore[DataplaneResponse]()
	projectStore   = NewSafeStore[Project]()
)

func init() {
	defaultProject := &Project{
		ProjectName:     "Default Project",
		ProjectId:       "proj-ebc5ac7f430702aec8c57b",
		InstanceCount:   0,
		CreateTimeMilli: 1703745469000, // 2023-12-28T07:17:49Z
		Plan:            "Enterprise",
	}
	projectStore.Set(defaultProject.ProjectId, defaultProject)
}

type IStore[T any] interface {
	Get(key string) *T
	Set(key string, value *T)
}

type safeStore[T any] struct {
	store sync.Map
}

func NewSafeStore[T any]() *safeStore[T] {
	return &safeStore[T]{
		store: sync.Map{},
	}
}

func (s *safeStore[T]) Get(key string) *T {
	value, ok := s.store.Load(key)
	if !ok {
		return nil
	}
	return value.(*T)
}

func (s *safeStore[T]) Set(key string, value *T) {
	s.store.Store(key, value)
}

func (s *safeStore[T]) GetAll() []*T {
	var results []*T
	s.store.Range(func(key, value interface{}) bool {
		results = append(results, value.(*T))
		return true
	})
	return results
}

func (s *safeStore[T]) Delete(key string) {
	s.store.Delete(key)
}
