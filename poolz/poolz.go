package poolz

import (
	"sync"
	"time"
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
	wrapper *openerCloserWrapper
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

	// Return count of Poolable available and in use.
	Count() (int, int)
}

type openerCloserWrapper struct {
	wrapped Poolable

	onClose   func() error
	startIdle *time.Time
	creation  *time.Time
	outdated  bool
}

func (o openerCloserWrapper) Get() Poolable {
	return o.wrapped
}

func (o openerCloserWrapper) Close() error {
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
	available      []*openerCloserWrapper
	cleanerRunning bool
}

func (p *basicPool[K]) startCleaner() {
	if p.cleanerRunning {
		return
	}
	p.cleanerRunning = true
	for len(p.available)+p.inUse > 0 {
		p.Lock()
		//var toClose []*openerCloserWrapper
		for _, w := range p.available {
			if p.maxIdleTime != nil && time.Since(*w.startIdle) > *p.maxIdleTime ||
				p.maxLifeTime != nil && time.Since(*w.creation) > *p.maxLifeTime {
				w.outdated = true
				//toClose = append(toClose, w)
			}
		}

		// for _, w := range toClose {
		// 	w.wrapped.Close()
		// 	p.available = collectionz.DeleteFast(p.available, w)
		// }

		p.Unlock()
		time.Sleep(time.Second)
	}
	p.cleanerRunning = false
}

func (p *basicPool[K]) Count() (int, int) {
	p.Lock()
	defer p.Unlock()

	return len(p.available), p.inUse
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
	defer func() {
		go p.startCleaner()
	}()

	for {
		p.Lock()
		if len(p.available) > 0 {
			// Take first OpenerCloser available
			first := p.available[0]
			p.inUse++
			p.available = p.available[1:]
			opened := (first.wrapped).(K)
			p.Unlock()
			return opened, nil
		} else if p.inUse <= p.maxOpen {
			// Build new OpenerCloser
			newOc, err := p.factory()
			if err != nil {
				p.Unlock()
				return o, err
			}
			err = newOc.Open()
			if err != nil {
				p.Unlock()
				return o, err
			}
			now := time.Now()
			wrapper := &openerCloserWrapper{wrapped: newOc, creation: &now}
			wrapper.onClose = func() error {
				if !wrapper.outdated {
					p.available = append(p.available, wrapper)
					now := time.Now()
					wrapper.startIdle = &now
				}
				p.inUse--
				return nil
			}
			p.inUse++
			opened := (wrapper.wrapped).(K)
			spy := PoolCloser{wrapper: wrapper}
			opened.SetPoolCloser(spy)
			p.Unlock()
			return opened, err
		}
		p.Unlock()
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
