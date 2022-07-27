package concurrent

import (
	"fmt"
	"sync"

	"mby.fr/utils/errorz"
)

// For naming ideas cf: https://docs.oracle.com/javase/8/docs/api/java/util/function/package-summary.html

type Consumer[I any] func(I) error 			// 1 input => no output
type Supplier[O any] func() (O, error) 		// No input => 1 output
//type Predicate[I any] func(I) (bool, error) // 1 input => bool output same as Function
type Function[I, O any] func(I) (O, error) 	// 1 input => 1 output

type ErrorConsumer func(error)

func consume[F Consumer[I]|Function[I, O], I, O any](f F, in I) (O, error) {
	var out O
	var err error

	switch v := any(f).(type) {
	case nil: 
		err = fmt.Errorf("p is nil !")
	case Consumer[I]: 
		err = v(in)
	case Function[I, O]: 
		out, err = v(in)
	default:
		err = fmt.Errorf("Not supported %T type !", v)
	}
	
	return out, err
}

func Run[I, O any](f Function[I, O], inputs ...I) (*sync.WaitGroup, chan O, chan error) {
	var wg sync.WaitGroup
	outputs := make(chan O, len(inputs))
	errors := make(chan error, len(inputs))
	for _, in := range inputs {
		wg.Add(1)
		go func(i I) {
			defer wg.Done()
			out, err := f(i)
			if err != nil {
				errors <- err
			} else {
				outputs <- out
			}
		}(in)
	}
	//wg.Wait()
	return &wg, outputs, errors
}

func RunWaiting[I, O any](f Function[I, O], inputs ...I) (chan O, error) {
	wg, outputs, errors := Run(f, inputs...)
	wg.Wait()
	err := errorz.ConsumedAggregated(errors)
	return outputs, err.Return()
}
