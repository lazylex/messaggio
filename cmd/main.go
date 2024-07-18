package main

import (
	"github.com/lazylex/messaggio/internal/config"
)

func main() {
	_ = config.MustLoad()
}
