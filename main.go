package main

import (
	"database/sql"
	"fmt"
	"gator/internal/command"
	"gator/internal/config"
    "gator/internal/database"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: gator <command> [args...]")
        os.Exit(1)
    }

    cfg, err := config.Read()
    if err != nil {
        log.Fatalf("unable to read config %v", err) 
    }


    db, err := sql.Open("postgres", cfg.Db_url)
    if err != nil {
        log.Fatalf("unable to open db connection, error: %v", err)
    }

    var state command.State
    state.Cfg = &cfg
    state.Db = database.New(db)

    cmds := command.GetCommands()

    args := os.Args
    if len(args) == 0 {
        log.Fatal("No arguments found")
    }

    cmd := command.GetCommand(args[1], args[2:])

    err = cmds.CommandMap[cmd.Name](&state, cmd)
    if err != nil {
        log.Fatalf("Failed to execute command %s, Error: %v", cmd.Name, err)
    }


    cfg, err = config.Read()
    if err != nil {
        log.Fatalf("unable to read config %v", err)
    }
    fmt.Printf("Config: \n  Username: %s,\n  Db_Url: %s\n", cfg.Username, cfg.Db_url) 
}
