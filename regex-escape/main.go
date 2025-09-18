package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	// Read a single line from stdin
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return
	}
	input := scanner.Text()

	// Escape it for regex
	escaped := regexp.QuoteMeta(input)

	// Escape again for Go string literal (double the backslashes)
	goLiteral := strings.ReplaceAll(escaped, `\`, `\\`)

	fmt.Println(goLiteral)
}
