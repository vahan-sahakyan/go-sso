package main

import (
  "fmt"
  "sso/internal/config"
)

func main() {
  cfg := config.MustLoad()

  fmt.Println(cfg.ToString())

  // TODO: initialize logger

  // TODO: initialize app

  // TODO: launch gRPC-server apps

}
