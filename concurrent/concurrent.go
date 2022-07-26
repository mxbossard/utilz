package concurrent

import (
	"sync"
)

type Processor[I, O any] func(I) (O, error)
type OnError func(error)

func Run[I, O any](p Processor[I, O], inputs ...I) (sync.WaitGroup, chan O, chan error) {
	var wg sync.WaitGroup
	outputs := make(chan O, len(inputs))
	errors := make(chan error, len(inputs))
	for _, in := range inputs {
		wg.Add(1)
		go func(i I) {
			defer wg.Done()
			out, err := p(i)
			if err != nil {
				errors <- err
			} else {
				outputs <- out
			}
		}(in)
	}
	return wg, outputs, errors
}

func RunWaiting[I, O any](p Processor[I, O], inputs ...I) (chan O, chan error) {
	/*var wg sync.WaitGroup
	outputs := make(chan O, len(inputs))
	errors := make(chan error, len(inputs))
	for _, in := range inputs {
		wg.Add(1)
		go func(i I) {
			defer wg.Done()
			out, err := p(i)
			if err != nil {
				errors <- err
			} else {
				outputs <- out
			}
		}(in)
	}*/
	wg, outputs, errors := Run(p, inputs...)
	wg.Wait()
	return outputs, errors
}
