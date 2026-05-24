package command

import (
	"context"
	"errors"
	"gator/internal/config"
	"gator/internal/database"
	"log"
    "time"

	"github.com/google/uuid"
)

type State struct {
    Db  *database.Queries
    Cfg *config.Config
}

type Command struct {
    Name string
    Args []string
}

type Commands struct {
    CommandMap map[string]func(*State, Command) error
}

func (cmds *Commands) run(s *State, cmd Command) error {
    err := cmds.CommandMap[cmd.Name](s, cmd)
    if err != nil {
        return err
    }

    return nil
}

func (cmds *Commands) register(name string, f func(*State, Command) error) {
    cmds.CommandMap[name] = f
}

func handlerLogin(s *State, cmd Command) error {
    if len(cmd.Args) == 0 {
        return errors.New("no arguements found for command 'login'")
    }

    name := cmd.Args[0]

    user, _ := s.Db.GetUser(context.Background(), name)
    if user == (database.User{}) {
        return errors.New("User does not exists")
    }

    s.Cfg.SetUser(cmd.Args[0])
    log.Printf("User: %s, has logged in", s.Cfg.Username)

    return nil
}

func handleRegister(s *State, cmd Command) error {
    if len(cmd.Args) == 0 {
        return errors.New("no arguments found for command 'register'")
    }
    name := cmd.Args[0]

    existingUser, err := s.Db.GetUser(context.Background(), name)
    if existingUser != (database.User{}) {
        return errors.New("User " + name + " already exists")
    }

    var params database.CreateUserParams
    params.ID = uuid.New()
    params.CreatedAt = time.Now()
    params.UpdatedAt = time.Now()
    params.Name = name

    user, err := s.Db.CreateUser(context.Background(), params)
    if err != nil {
        return err
    }

    log.Printf("User %s was created with UUID: %v", user.Name, user.ID)

    s.Cfg.SetUser(user.Name)

    return nil
}

func GetCommands() Commands {
    var cmds Commands

    cmds.CommandMap = make(map[string]func(*State, Command) error)

    cmds.register("login", handlerLogin)
    cmds.register("register", handleRegister)
   

    return cmds
}

func GetCommand(name string, args []string) Command {
    var cmd Command
    cmd.Name = name
    cmd.Args = args

    return cmd
}
