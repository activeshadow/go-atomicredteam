package atomicredteam

import "fmt"

var Quiet bool

func Println(a ...interface{}) (int, error) {
	if Quiet {
		return 0, nil
	}

	return fmt.Println(a...)
}

func Printf(format string, a ...interface{}) (int, error) {
	if Quiet {
		return 0, nil
	}

	return fmt.Printf(format, a...)
}

func Print(a ...interface{}) (int, error) {
	if Quiet {
		return 0, nil
	}

	return fmt.Print(a...)
}
