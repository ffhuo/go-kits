package utils

import (
	"fmt"
	"os"
	"strings"
)

var (
	DefaultWriter = os.Stdout
)

func DebugPrint(format string, values ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(DefaultWriter, format, values...)
}
