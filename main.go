package main

import (
	"fmt"
	"gator/internal/config"
    "log"
)

func main() {
    cfg, err := config.Read()
    if err != nil {
        log.Fatalf("unable to read config %v", err) 
    }

    err = cfg.SetUser("Max")
    if err != nil {
        log.Fatalf("unable to set user %v", err)
    }

    cfg, err = config.Read()
    if err != nil {
        log.Fatalf("unable to read config %v", err)
    }
    fmt.Printf("Config: \n  Username: %s,\n  Db_Url: %s\n", cfg.Username, cfg.Db_url) 
}
