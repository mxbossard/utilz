package poolz

import (
	"sync"
	"time"

	"github.com/mxbossard/utilz/collectionz"
)

type Closer interface {
	Close() error
}

type Poolable interface {
	Open() error
	Close() error
	PoolClose() error
	SetPoolCloser(s PoolCloser)
}

type PoolCloser struct {
	wrapper *openerCloserWrapper2
}

func (s PoolCloser) PoolClose() error {
	return s.wrapper.Close()
}

type Pool[K Poolable] interface {
	//SetFatory(func() (OpenerCloser, error))
	SetMaxOpen(int)
	SetMaxIdle(int)
	SetMaxIdleTime(time.Duration)
	SetMaxLifeTime(time.Duration)

	Open() (K, error)
}

type openerCloserWrapper2 struct {
	wrapped Poolable

	onClose   func() error
	startIdle *time.Time
	creation  *time.Time
}

func (o openerCloserWrapper2) Get() Poolable {
	return o.wrapped
}

func (o openerCloserWrapper2) Close() error {
	return o.onClose()
}

type basicPool[K Poolable] struct {
	*sync.Mutex

	factory        func() (K, error)
	maxOpen        int
	maxIdle        int
	maxIdleTime    *time.Duration
	maxLifeTime    *time.Duration
	inUse          int
	available      []*openerCloserWrapper2
	cleanerRunning bool
}

func (p *basicPool[K]) startCleaner() {
	if p.cleanerRunning {
		return
	}
	p.cleanerRunning = true
	for len(p.available)+p.inUse > 0 {
		p.Lock()
		var toClose []*openerCloserWrapper2
		for _, w := range p.available {
			if p.maxIdleTime != nil && time.Since(*w.startIdle) > *p.maxIdleTime ||
				p.maxLifeTime != nil && time.Since(*w.creation) > *p.maxLifeTime {
				toClose = append(toClose, w)
			}
		}

		for _, w := range toClose {
			w.wrapped.Close()
			p.available = collectionz.DeleteFast(p.available, w)
		}

		p.Unlock()
		time.Sleep(time.Second)
	}
	p.cleanerRunning = false
}

func (p *basicPool[K]) SetMaxOpen(n int) {
	p.maxOpen = n
}

func (p *basicPool[K]) SetMaxIdle(n int) {
	p.maxIdle = n
}

func (p *basicPool[K]) SetMaxIdleTime(d time.Duration) {
	p.maxIdleTime = &d
}

func (p *basicPool[K]) SetMaxLifeTime(d time.Duration) {
	p.maxLifeTime = &d
}

func (p *basicPool[K]) Open() (o K, err error) {
	p.Lock()
	defer p.Unlock()

	defer func() {
		go p.startCleaner()
	}()

	for {
		if len(p.available) > 0 {
			// Take first OpenerCloser available
			first := p.available[0]
			p.inUse++
			p.available = p.available[1:]
			opened := (first.wrapped).(K)
			return opened, nil
		} else if p.inUse <= p.maxOpen {
			// Build new OpenerCloser
			newOc, err := p.factory()
			if err != nil {
				return o, err
			}
			err = newOc.Open()
			if err != nil {
				return o, err
			}
			now := time.Now()
			wrapper := &openerCloserWrapper2{wrapped: newOc, creation: &now}
			wrapper.onClose = func() error {
				p.available = append(p.available, wrapper)
				p.inUse--
				now := time.Now()
				wrapper.startIdle = &now
				return nil
			}
			p.inUse++
			opened := (wrapper.wrapped).(K)
			spy := PoolCloser{wrapper: wrapper}
			opened.SetPoolCloser(spy)
			return opened, err
		}
		time.Sleep(100 * time.Microsecond)
	}
}

func NewBasic[K Poolable](maxSize int, factory func() (K, error)) *basicPool[K] {
	return &basicPool[K]{
		Mutex:   &sync.Mutex{},
		maxOpen: maxSize,
		factory: factory,
	}
}

func New[K Poolable](maxSize int, factory func() (K, error)) *basicPool[K] {
	return NewBasic(maxSize, factory)
}
