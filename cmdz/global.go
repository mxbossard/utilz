package cmdz

import (
	//"bufio"
	"context"
	"fmt"

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

func BlockParallelRunAll(forkCount int, execs ...Executer) ([]int, error) {
	p := AsyncRunAll(execs...)
	br, err := WaitAllResults(p)
	if err != nil {
		return nil, err
	}

	return *br, nil
}

func BlockParallel(forkCount int, execs ...Executer) error {
	statuses, err := BlockParallelRunAll(forkCount, execs...)

	if err != nil {
		return err
	}
	if Failed(statuses...) {
		errorMessages := ""
		for idx, status := range statuses {
			if status > 0 {
				//stdout := execs[idx].StdoutRecord.String()
				execErr := execs[idx].ReportError()
				errorMessages += fmt.Sprintf("%s\n", execErr)
			}
		}
		return fmt.Errorf("Encountered some parallel execution failure: \n%s", errorMessages)
	}
	return nil
}

func BlockSerial(execs ...Executer) error {
	for _, exec := range execs {
		status, err := exec.BlockRun()
		if err != nil {
			return err
		}
		if status > 0 {
			execErr := exec.ReportError()
			return fmt.Errorf("Encountered some sequential execution failure: \n%s", execErr)
		}
	}
	return nil
}
