package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
)

const (
	GREEN  = "Green"
	RESET  = "Reset"
	RED    = "Red"
	YELLOW = "Yellow"

	ANSIRESET  string = "\x1b[0m"
	ANSIRED    string = "\x1b[31m"
	ANSIGREEN  string = "\x1b[92m"
	ANSIYELLOW string = "\x1b[33m"
)

var colormap map[string]string = map[string]string{RESET: ANSIRESET, RED: ANSIRED, GREEN: ANSIGREEN, YELLOW: ANSIYELLOW}

func Errhandle_Exit(err error, reason string) {
	var file string
	var line int
	fmt.Printf("%s:", reason)
	if err != nil {
		SetConsoleColor(RED)
		_, file, line, _ = runtime.Caller(1)
		fmt.Printf("\t%s %d\t failed: %v\n", file, line, err)
		SetConsoleColor(RESET)
		os.Exit(1)
	} else {
		SetConsoleColor(GREEN)
		fmt.Printf("\t successful.\n")
		SetConsoleColor(RESET)
	}
}

func Errhandle_Log(err error, reason string) {
	var file string
	var line int
	fmt.Printf("%s:", reason)
	if err != nil {
		SetConsoleColor(YELLOW)
		_, file, line, _ = runtime.Caller(1)
		fmt.Printf("\t%s %d\t failed: %v\n", file, line, err)
		SetConsoleColor(RESET)
	} else {
		SetConsoleColor(GREEN)
		fmt.Printf("\t successful.\n")
		SetConsoleColor(RESET)
	}
}

func Addcolorpair(key, ansicode string) {
	switch len(key) {
	case 0:
		_, fn, line, _ := runtime.Caller(1)
		log.Printf("%sline %d, function %s%s", ANSIRED, line, fn, ANSIRESET)
		log.Printf("length 0 color key (\"consoleio.go\" line 18\n")
	default:
		colormap[key] = ansicode
	}
}

func SetConsoleColor(color string) {
	for key, value := range colormap {
		if key == color {
			fmt.Printf("%s", value)
		}
	}
}
