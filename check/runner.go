package check

import "fmt"

type Runner struct {
	checksRun    int
	checksPassed int
}

type Check func() bool

func NewRunner() Runner {
	r := Runner{
		checksRun:    0,
		checksPassed: 0,
	}

	return r
}

func (r *Runner) RunCheck(condition string, f Check) error {
	passed := f()
	if passed {
		fmt.Printf("PASS: %s\n", condition)
	} else {
		fmt.Printf("FAIL: %s\n", condition)
	}

	r.registerResult(passed)

	return nil
}

func (r *Runner) Summary() {
	checksFailed := r.checksRun - r.checksPassed
	fmt.Printf("Checks completed: [%d], Passed: [%d], Failed: [%d]\n", r.checksRun, r.checksPassed, checksFailed)
}

func (r *Runner) registerResult(result bool) {
	r.checksRun++
	if result == true {
		r.checksPassed++
	}
}
