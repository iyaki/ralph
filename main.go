package main

import (
"fmt"
"os"
)

func main() {
cmd := NewRalphCommand()
if err := cmd.Execute(); err != nil {
fmt.Fprintf(os.Stderr, "Error: %v\n", err)
os.Exit(1)
}
}
