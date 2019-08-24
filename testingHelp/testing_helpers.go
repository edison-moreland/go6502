/*
Original: https://github.com/benbjohnson/testing

Changes:
5/16/2018 - Edison Moreland, added type information to errors and support for pkg/errors stacktrace
8/24/2019 - Edison Moreland, Exported functions
*/

package testingHelp

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// Assert fails the test if the condition is false.
func Assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// NotNil fails the test if an err is not nil.
func NotNil(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %+v\033[39m\n\n", filepath.Base(file), line, err)
		tb.FailNow()
	}
}

// Equals fails the test if exp is not equal to act.
func Equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%[1]s:%[2]d:\n\n\texp: (%[3]T)%#[3]v\n\n\tgot: (%[4]T)%#[4]v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}
