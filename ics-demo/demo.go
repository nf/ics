package main

import (
	"encoding/json"
	"fmt"
	"github.com/nf/ics"
	"os"
)

func main() {
	c, err := ics.Decode(os.Stdin)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	b, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("%s\n", b)
}
