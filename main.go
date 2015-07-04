package main

import (
    "fmt"
    "github.com/FogCreek/mini"
)

func fatal(v interface{}) {
	fmt.Println(v)
}

func chk(err error) {
	if err != nil {
		fatal(err)
	}
}

func params() string {
	cfg, err := mini.LoadConfiguration(".pulseconfigrc")

    chk(err)

    fmt.Printf("params")


	info := fmt.Sprintf("db=%s",
		cfg.String("db", "127.0.0.1"),
	)
	return info
}

func main() {
    fmt.Printf("hello, world\n")
}

