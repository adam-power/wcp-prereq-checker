package main

import (
	"github.com/adam-power/wcp-prereq-checker/check"
	"github.com/adam-power/wcp-prereq-checker/iaas"
)

func main() {
	iaas.RunIaaSChecks()
	check.Summary()
}
