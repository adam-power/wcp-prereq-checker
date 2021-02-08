package check

import "fmt"

var (
	checksRun    int
	checksPassed int
)

func init() {
	checksRun = 0
	checksPassed = 0
}
func Summary() {
	checksFailed := checksRun - checksPassed
	fmt.Printf("Checks completed: [%d], Passed: [%d], Failed: [%d]\n", checksRun, checksPassed, checksFailed)
}

func RegisterResult(check string, result bool, err error) {
	checksRun++
	if result == true {
		checksPassed++
		fmt.Printf("PASS: %s\n", check)
	} else {
		fmt.Printf("FAIL: %s\n", check)
	}

	if err != nil {
		fmt.Printf("      -> ERROR: %v\n", err)
	}
}
