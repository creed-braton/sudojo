package main

import (
	"fmt"
	"os"
)

func main() {
}

func envOrPanic(key string) string {
	val := os.Getenv(key)
	if len(val) < 1 {
		panic(fmt.Errorf("env variable '%s' missing", key))
	}
	return val
}
