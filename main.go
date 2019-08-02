package main

import (
	"fmt"
	"os"
	"testGorillaMux/server"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:", os.Args[0], "[config_path]")
		return
	}
	s := new(server.Server)

	err := s.Init(os.Args[1])
	if err != nil {
		panic(err)
	}

	s.Run()
}
