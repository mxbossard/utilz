package poolz

import (
	"sync"
	"time"

	"github.com/mxbossard/utilz/collectionz"
)

type OpenerCloser0 interface {
	Open() error
	Close() error
}

type Pool0[K OpenerCloser0] interface {
	//SetFatory(func() (OpenerCloser, error))
	SetMaxOpen(int)
	SetMaxIdle(int)
	SetMaxIdleTime(time.Duration)
	SetMaxLifeTime(time.Duration)

	Open() (K, error)
}

type openerCloserWrapper0 struct {
	wrapped OpenerCloser0

	onClose   func() error
	startIdle *time.Time
	creation  *time.Time
}

func (o openerCloserWrapper0) Get() OpenerCloser0 {
	return o.wrapped
}

func (o openerCloserWrapper0) Close() error {
	return o.onClose()
}

type basicPool0[K OpenerCloser0] struct {
	*sync.Mutex

	factory        func() (K, error)
	maxOpen        int
	maxIdle        int
	maxIdleTime    *time.Duration
	maxLifeTime    *time.Duration
	inUse          int
	available      []*openerCloserWrapper0
	cleanerRunning bool
}

func (p *basicPool0[K]) startCleaner() {
	if p.cleanerRunning {
		return
	}
	p.cleanerRunning = true
	for len(p.available)+p.inUse > 0 {
		p.Lock()
		var toClose []*openerCloserWrapper0
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

func (p *basicPool0[K]) SetMaxOpen(n int) {
	p.maxOpen = n
}

func (p *basicPool0[K]) SetMaxIdle(n int) {
	p.maxIdle = n
}

func (p *basicPool0[K]) SetMaxIdleTime(d time.Duration) {
	p.maxIdleTime = &d
}

func (p *basicPool0[K]) SetMaxLifeTime(d time.Duration) {
	p.maxLifeTime = &d
}

func (p *basicPool0[K]) Open() (o K, err error) {
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
			opened := first.wrapped.(K)
			return opened, nil
		} else if p.inUse <= p.maxOpen {
			// Build new OpenerCloser
			newOc, err := p.factory()
			if err != nil {
				return o, err
			}
			err = (newOc).Open()
			if err != nil {
				return o, err
			}
			now := time.Now()
			wrapper := openerCloserWrapper0{wrapped: newOc, creation: &now}
			wrapper.onClose = func() error {
				p.available = append(p.available, &wrapper)
				p.inUse--
				now := time.Now()
				wrapper.startIdle = &now
				return nil
			}
			p.inUse++
			opened := wrapper.wrapped.(K)
			return opened, err
		}
		time.Sleep(100 * time.Microsecond)
	}
}

func NewBasic0[K OpenerCloser0](maxSize int, factory func() (K, error)) *basicPool0[K] {
	return &basicPool0[K]{
		Mutex:   &sync.Mutex{},
		maxOpen: maxSize,
		factory: factory,
	}
}

func New0[K OpenerCloser0](maxSize int, factory func() (K, error)) *basicPool0[K] {
	return NewBasic0(maxSize, factory)
}
