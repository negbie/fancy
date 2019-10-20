package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		lm, err := scanLine(s.Bytes())
		if err != nil {
			return
		}
		fmt.Fprintf(os.Stderr, lm.String(), lm.Valid())

	}
}
