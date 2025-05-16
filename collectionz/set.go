package collectionz

import "sync"

/*

/!\ WARNING: first naive implem not used nor tested !!! /!\

*/

type Set[K comparable] interface {
	Add(item K)
	Remove(item K)
	Len() int
	Clear()
	Iterate() []K
}

type basicSet[K comparable] struct {
	*sync.Mutex
	repo map[K]bool
}

func (s *basicSet[K]) Add(item K) {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.repo[item]; !ok {
		s.repo[item] = true
	}
}

func (s *basicSet[K]) Remove(item K) {
	s.Lock()
	defer s.Unlock()
	delete(s.repo, item)
}

func (s basicSet[K]) Len() int {
	s.Lock()
	defer s.Unlock()
	return len(s.repo)
}

func (s *basicSet[K]) Clear() {
	s.Lock()
	defer s.Unlock()
	s.repo = make(map[K]bool, 8)
}

func (s *basicSet[K]) Iterate() []K {
	s.Lock()
	defer s.Unlock()
	return Keys(s.repo)
}

func NewBasicSet[K comparable]() *basicSet[K] {
	return &basicSet[K]{
		Mutex: &sync.Mutex{},
	}
}

func NewSet[K comparable]() *basicSet[K] {
	return NewBasicSet[K]()
}
