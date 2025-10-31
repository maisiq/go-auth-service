package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/maisiq/go-auth-service/cmd/wire"
	xhttp "github.com/maisiq/go-auth-service/internal/transport/http"
)

func main() {
	_ = context.Background()

	cfgName, ok := os.LookupEnv("CONFIG_FILENAME")
	if !ok {
		panic("config path is not provided. Set CONFIG_FILENAME env")
	}

	di := BuildContainer(fmt.Sprintf("./configs/%s", cfgName))
	defer di.ShutdownResources()

loop:
	for {
		router := wire.Get[*gin.Engine](di)
		srv := wire.Get[*xhttp.Server](di)
		srv.RunWithHandler(router)

		// rebuild container on config change
		select {
		case <-di.Rebuild():
			continue loop
		default:
			break loop
		}
	}
}
