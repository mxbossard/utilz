package promise

// Copied from: https://github.com/chebyrash/promise

import (
	"context"
	"fmt"
	"sync"
)

type Promise[T any] struct {
	value *T
	err   error
	ch    chan struct{}
	once  sync.Once
}

func New[T any](executor func(resolve func(T), reject func(error))) *Promise[T] {
	if executor == nil {
		panic("missing executor")
	}

	p := &Promise[T]{
		value: nil,
		err:   nil,
		ch:    make(chan struct{}),
		once:  sync.Once{},
	}

	go func() {
		defer p.handlePanic()
		executor(p.resolve, p.reject)
	}()

	return p
}

func Then[A, B any](p *Promise[A], ctx context.Context, resolve func(A) B) *Promise[B] {
	return New(func(internalResolve func(B), reject func(error)) {
		result, err := p.Await(ctx)
		if err != nil {
			reject(err)
		} else {
			internalResolve(resolve(*result))
		}
	})
}

func Catch[T any](p *Promise[T], ctx context.Context, reject func(err error) error) *Promise[T] {
	return New(func(resolve func(T), internalReject func(error)) {
		result, err := p.Await(ctx)
		if err != nil {
			internalReject(reject(err))
		} else {
			resolve(*result)
		}
	})
}

func (p *Promise[T]) Await(ctx context.Context) (*T, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-p.ch:
		return p.value, p.err
	}
}

func (p *Promise[T]) resolve(value T) {
	p.once.Do(func() {
		p.value = &value
		close(p.ch)
	})
}

func (p *Promise[T]) reject(err error) {
	p.once.Do(func() {
		p.err = err
		close(p.ch)
	})
}

func (p *Promise[T]) handlePanic() {
	err := recover()
	if err == nil {
		return
	}

	switch v := err.(type) {
	case error:
		p.reject(v)
	default:
		p.reject(fmt.Errorf("%+v", v))
	}
}

// All resolves when all promises have resolved, or rejects immediately upon any of the promises rejecting
func All[T any](ctx context.Context, promises ...*Promise[T]) *Promise[[]T] {
	if len(promises) == 0 {
		panic("missing promises")
	}

	return New(func(resolve func([]T), reject func(error)) {
		resultsChan := make(chan tuple[T, int], len(promises))
		errsChan := make(chan error, len(promises))

		for idx, p := range promises {
			idx := idx
			_ = Then(p, ctx, func(data T) T {
				resultsChan <- tuple[T, int]{_1: data, _2: idx}
				return data
			})
			_ = Catch(p, ctx, func(err error) error {
				errsChan <- err
				return err
			})
		}

		results := make([]T, len(promises))
		for idx := 0; idx < len(promises); idx++ {
			select {
			case result := <-resultsChan:
				results[result._2] = result._1
			case err := <-errsChan:
				reject(err)
				return
			}
		}
		resolve(results)
	})
}

// Race resolves or rejects as soon as any one of the promises resolves or rejects
func Race[T any](ctx context.Context, promises ...*Promise[T]) *Promise[T] {
	if len(promises) == 0 {
		panic("missing promises")
	}

	return New(func(resolve func(T), reject func(error)) {
		valsChan := make(chan T, len(promises))
		errsChan := make(chan error, len(promises))

		for _, p := range promises {
			_ = Then(p, ctx, func(data T) T {
				valsChan <- data
				return data
			})
			_ = Catch(p, ctx, func(err error) error {
				errsChan <- err
				return err
			})
		}

		select {
		case val := <-valsChan:
			resolve(val)
		case err := <-errsChan:
			reject(err)
		}
	})
}

type BestResults[T any] struct {
	Results []T
	Errors  []error
}

func (br *BestResults[T]) init(size int) {
	br.Results = make([]T, size)
	br.Errors = make([]error, size)
}

func (br *BestResults[T]) Len() int {
	return len(br.Results)
}

func (br BestResults[T]) AddResult(idx int, result T) {
	br.Results[idx] = result
}

func (br BestResults[T]) AddError(idx int, err error) {
	br.Errors[idx] = err
}

func (br BestResults[T]) DidError() bool {
	for _, err := range br.Errors {
		if err != nil {
			return true
		}
	}
	return false
}

func (br BestResults[T]) Result(idx int) (T, error) {
	return br.Results[idx], br.Errors[idx]
}

// Best resolves when all promises have resolved or rejected. return ok = true when no error
func Best[T any](ctx context.Context, promises ...*Promise[T]) *Promise[BestResults[T]] {
	if len(promises) == 0 {
		panic("missing promises")
	}

	return New(func(resolve func(BestResults[T]), reject func(error)) {
		resultsChan := make(chan tuple[T, int], len(promises))
		errsChan := make(chan tuple[error, int], len(promises))

		for idx, p := range promises {
			idx := idx
			_ = Then(p, ctx, func(data T) T {
				resultsChan <- tuple[T, int]{_1: data, _2: idx}
				return data
			})
			_ = Catch(p, ctx, func(err error) error {
				errsChan <- tuple[error, int]{_1: err, _2: idx}
				return err
			})
		}

		results := BestResults[T]{}
		results.init(len(promises))
		for idx := 0; idx < len(promises); idx++ {
			select {
			case result := <-resultsChan:
				results.AddResult(result._2, result._1)
			case err := <-errsChan:
				results.AddError(err._2, err._1)
			}
		}
		resolve(results)
	})
}

type tuple[T1, T2 any] struct {
	_1 T1
	_2 T2
}
