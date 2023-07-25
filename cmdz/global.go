package cmdz

import (
	//"bufio"
	"context"
	//"fmt"

	"mby.fr/utils/promise"
)

func AsyncRunAll(execs ...Executer) *execsPromise {
	var promises []*promise.Promise[int]
	for _, e := range execs {
		p := e.AsyncRun()
		promises = append(promises, p)
	}

	ctx := context.Background()
	p := promise.All[int](ctx, promises...)
	return p
}

func WaitAllResults(p *execsPromise) (*[]int, error) {
	ctx := context.Background()
	return p.Await(ctx)
}

func AsyncRunBest(execs ...Executer) *promise.Promise[promise.BestResults[int]] {
	var promises []*promise.Promise[int]
	for _, e := range execs {
		p := e.AsyncRun()
		promises = append(promises, p)
	}

	ctx := context.Background()
	p := promise.Best[int](ctx, promises...)
	return p
}

func WaitBestResults(p *promise.Promise[promise.BestResults[int]]) (*promise.BestResults[int], error) {
	ctx := context.Background()
	br, err := p.Await(ctx)
	if err != nil {
		return nil, err
	}
	if br.DidError() {
		return nil, br.AggError()
	}
	return br, err
}

func Failed(resultCodes ...int) bool {
	for _, rc := range resultCodes {
		if rc != 0 {
			return true
		}
	}
	return false
}

func Succeed(resultCodes ...int) bool {
	return !Failed(resultCodes...)
}

func blockParallelRunAll(forkCount int, execs ...Executer) ([]int, error) {
	p := AsyncRunAll(execs...)
	statuses, err := WaitAllResults(p)
	if err != nil {
		return nil, err
	}

	return *statuses, nil
}

func blockParallelRunBest(forkCount int, execs ...Executer) ([]int, error) {
	p := AsyncRunBest(execs...)
	br, err := WaitBestResults(p)
	if err != nil {
		return nil, err
	}

	return br.Results, nil
}

func blockParallel(failFast bool, forkCount int, execs ...Executer) (status int, err error) {
	if failFast {
		for _, exec := range execs {
			if c, ok := exec.(*cmdz); ok {
				c.errorOnFailure = true
			}
		}
	}
	statuses, err := blockParallelRunAll(forkCount, execs...)
	if err != nil {
		if f, ok := err.(failure); failFast && ok {
			return f.Rc, nil
		}
		return -1, err
	}

	if Failed(statuses...) {
		// Return first failure
		//errorMessages := ""
		for _, s := range statuses {
			if s > 0 {
				//execErr := execs[idx].ReportError()
				//errorMessages += fmt.Sprintf("%s\n", execErr)
				return s, nil
			}
		}
		//return fmt.Errorf("Encountered some parallel execution failure: \n%s", errorMessages)
	}
	return
}

func blockSerial(failFast bool, execs ...Executer) (status int, err error) {
	for _, exec := range execs {
		status, err = exec.BlockRun()
		if err != nil {
			return -1, err
		}

		if failFast && status > 0 {
			return
		}
	}
	return
}

func blockOr(execs ...Executer) (status int, err error) {
	for _, exec := range execs {
		status, err = exec.BlockRun()
		if err != nil {
			return -1, err
		}

		if status == 0 {
			return
		}
	}
	return
}
