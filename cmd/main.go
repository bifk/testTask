package main

import (
	"fmt"
	"github.com/bifk/testTask/internal/config"
	"log"
)

func main() {
	con, err := config.Load()

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(con)
}
