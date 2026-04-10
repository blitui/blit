package btest

import "fmt"

// fmtSprintf and fmtSprintln are tiny indirection helpers so the
// capturingT in snapshot_test.go doesn't trip go vet's printf analyzer on
// its own wrapper methods.
func fmtSprintf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func fmtSprintln(args ...interface{}) string {
	return fmt.Sprintln(args...)
}
