package utils

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Logger struct {
	tag string
}

func NewLogger(tag string) Logger {
	var logger Logger
	logger.tag = tag
	return logger
}

func (logger Logger) Debug(values ...interface{}) {
	var datas []string
	for _, value := range values {
		datas = append(datas, fmt.Sprintf("%#v", value))
	}
	_, filename, line, _ := runtime.Caller(1)
	_, name := filepath.Split(filename)
	fmt.Printf("%s \033[34m[D] [%s:%d]\033[32m%s\033[0m %s\n", time.Now().Format(time.RFC3339), name, line, logger.tag, strings.Join(datas, " "))
}

func (logger Logger) Check(err error, backLevel ...int) (ok bool) {
	level := 1
	if len(backLevel) != 0 {
		level = backLevel[0]
	}
	if err != nil {
		_, filename, line, _ := runtime.Caller(level)
		_, name := filepath.Split(filename)
		fmt.Printf("%s \033[31m[E] [%s:%d]\033[32m%s\033[0m %s\n", time.Now().Format(time.RFC3339), name, line, logger.tag, err.Error())
		return false
	}
	return true
}

func (logger Logger) PrintErr(values ...interface{}) {
	for _, v := range values {
		if err, ok := v.(error); ok {
			logger.Check(err, 2)
		}
	}
}

func Check(err error, backLevel ...int) (ok bool) {
	level := 1
	if len(backLevel) != 0 {
		level = backLevel[0]
	}
	if err != nil {
		_, filename, line, _ := runtime.Caller(level)
		_, name := filepath.Split(filename)
		fmt.Printf("%s \033[31m[E] [%s:%d]\033[0m %s\n", time.Now().Format(time.RFC3339), name, line, err.Error())
		return false
	}
	return true
}

func Debug(values ...interface{}) {
	var datas []string
	for _, value := range values {
		datas = append(datas, fmt.Sprintf("%#v", value))
	}
	_, filename, line, _ := runtime.Caller(1)
	_, name := filepath.Split(filename)
	fmt.Printf("%s \033[34m[D] [%s:%d]\033[0m %s\n", time.Now().Format(time.RFC3339), name, line, strings.Join(datas, " "))
}

func PrintErr(values ...interface{}) {
	for _, v := range values {
		if err, ok := v.(error); ok {
			Check(err, 2)
		}
	}
}
