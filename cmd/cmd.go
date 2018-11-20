package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"io/ioutil"
)

func FlagPlatform(v *string) {
	d, err := ioutil.TempDir("", "platform")
	if err != nil {
		panic(err)
	}
	flag.StringVar(v, "platform", d, "platform directory")
}

func FlagCache(v *string) {
	d, err := ioutil.TempDir("", "cache")
	if err != nil {
		panic(err)
	}
	flag.StringVar(v, "cache", d, "cache directory")
}

func FlagLaunch(v *string) {
	d, err := ioutil.TempDir("", "launch")
	if err != nil {
		panic(err)
	}
	flag.StringVar(v, "launch", d, "launch directory")
}

func FlagBuildpack(v *string) {
	d, err := ioutil.TempDir("", "buildpack")
	if err != nil {
		panic(err)
	}
	flag.StringVar(v, "buildpack", d, "buildpack directory for this buildpack")
}

const (
	CodeFailed      = 1
	CodeInvalidArgs = iota + 2
)

type ErrorFail struct {
	Err    error
	Code   int
	Action []string
}

func (e *ErrorFail) Error() string {
	message := "failed to " + strings.Join(e.Action, " ")
	if e.Err == nil {
		return message
	}
	return fmt.Sprintf("%s: %s", message, e.Err)
}

func FailCode(code int, action ...string) error {
	return FailErrCode(nil, code, action...)
}

func FailErr(err error, action ...string) error {
	code := CodeFailed
	if err, ok := err.(*ErrorFail); ok {
		code = err.Code
	}
	return FailErrCode(err, code, action...)
}

func FailErrCode(err error, code int, action ...string) error {
	return &ErrorFail{Err: err, Code: code, Action: action}
}

func Exit(err error) {
	if err == nil {
		os.Exit(0)
	}
	log.Printf("Error: %s\n", err)
	if err, ok := err.(*ErrorFail); ok {
		os.Exit(err.Code)
	}
	os.Exit(CodeFailed)
}