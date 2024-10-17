package main

import (
	"fmt"

	"github.com/chistyakoviv/converter/internal/config"
)

func main() {
	config := config.MustLoad()

	fmt.Printf("config: %+v\n", config)
}
