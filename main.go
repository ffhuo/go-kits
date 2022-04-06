package main

import (
	"context"

	"github.com/ffhuo/go-kits/logger"
)

func main() {
	log, err := logger.New(logger.WithTimeLayout("2006-01-02 15:04:05"))
	if err != nil {
		panic(err)
	}
	log.Info(context.Background(), "hello world")
}
