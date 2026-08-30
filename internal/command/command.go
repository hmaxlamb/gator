package command

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gator/internal/config"
	"gator/internal/database"
	"gator/internal/rss"
	"time"
	"strconv"

	"github.com/google/uuid"
	"github.com/lib/pq"
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
	fmt.Printf("User: %s, has logged in\n", s.Cfg.Username)

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

	fmt.Printf("User %s was created with UUID: %v\n", user.Name, user.ID)

	s.Cfg.SetUser(user.Name)

	return nil
}

func handleReset(s *State, cmd Command) error {
	if len(cmd.Args) > 0 {
		return errors.New("Too many arguments for command")
	}

	err := s.Db.ResetUsers(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func handleUsers(s *State, cmd Command) error {
	if len(cmd.Args) > 0 {
		return errors.New("Too many arguments for command")
	}

	currentUsername := s.Cfg.Username

	users, err := s.Db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.Name == currentUsername {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}

	return nil
}

func handleAgg(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return errors.New("Too many arguments for command")
	}
	
	time_between, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return err
	}

	ticker := time.NewTicker(time_between)

	for ; ; <-ticker.C {
		err = scrapFeeds(s, user)
		if err != nil {
			return err
		}
	}
}

func scrapFeeds(s *State, user database.User) error {
	fmt.Printf("Getting Next Feed\n")

	feed, err := s.Db.GetNextFeedToFetch(context.Background(), user.ID)
	if err != nil {
		return err
	}

	var markParams database.MarkFeedFetchParams
	markParams.ID = feed.ID
	markParams.UpdatedAt = time.Now()

	err = s.Db.MarkFeedFetch(context.Background(), markParams)
	if err != nil {
		return err
	}

	rssFeed, err := rss.FetchFeed(context.Background(), feed.Url)
	if err != nil {
		return err
	}

	for _, feedItem := range rssFeed.Channel.Item {
		var postParams database.CreatePostParams
		postParams.ID = uuid.New()
		postParams.CreatedAt = time.Now()
		postParams.UpdatedAt = time.Now()
		postParams.Title = feedItem.Title
		postParams.Description = newNullString(feedItem.Description)
		postParams.PublishedAt, err = parsePublishedDate(feedItem.PubDate)
		if err != nil {
			fmt.Printf("Could Not Parse Pub DATE!!!! for feed item: %s, Time %s\n", feedItem.Title, feedItem.PubDate)
			fmt.Printf("Error %v\n", err)
			continue
		}
		postParams.FeedID = feed.ID

		err := s.Db.CreatePost(context.Background(), postParams)
		if checkUniqueError(err) {
			continue
		} else if err != nil {
			return err
		}
	}

	return nil
}

func newNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}

	return sql.NullString{
		String: s,
		Valid: true,
	}
}

func parsePublishedDate(dateString string) (time.Time, error) {
	t, err := time.Parse(time.RFC822, dateString)
	if err != nil {
		t, err = time.Parse(time.RFC1123, dateString)
		if err != nil {
			return time.Time{}, err
		}
	}

	return t, nil
}

func checkUniqueError(e error) bool {
	var pqErr *pq.Error
	if errors.As(e, &pqErr) {
		if pqErr.Code == "23505" {
			return true
		}
	}

	return false
}

func handleAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 2 {
		return errors.New("Wrong number of args for command, Required: 2")
	}

	var params database.CreateFeedParams
	params.ID = uuid.New()
	params.Name = cmd.Args[0]
	params.Url = cmd.Args[1]
	params.CreatedAt = time.Now()
	params.UpdatedAt = time.Now()

	params.UserID = user.ID

	feed, err := s.Db.CreateFeed(context.Background(), params)
	if err != nil {
		return err
	}

	var followParams database.CreateFeedFollowParams
	followParams.ID = uuid.New()
	followParams.UserID = user.ID
	followParams.FeedID = feed.ID
	followParams.CreatedAt = time.Now()
	followParams.UpdatedAt = time.Now()

	_, err = s.Db.CreateFeedFollow(context.Background(), followParams)
	if err != nil {
		return err
	}

	fmt.Printf("Feed Added:\nName: %s\nURL: %s", feed.Name, feed.Url)

	return nil
}

func handleFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return errors.New("Wrong number of args for command, Required: 1")
	}

	feedUrl := cmd.Args[0]

	var params database.CreateFeedFollowParams
	params.ID = uuid.New()
	params.CreatedAt = time.Now()
	params.UpdatedAt = time.Now()

	params.UserID = user.ID

	feed, err := s.Db.GetFeed(context.Background(), feedUrl)
	if err != nil {
		return err
	}
	params.FeedID = feed.ID

	feedFollow, err := s.Db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}

	fmt.Printf("Feed %s followed by user %s\n", feedFollow.FeedName, feedFollow.UserName)

	return nil
}

func handleFollowing(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 0 {
		return errors.New("No args allowed for command")
	}

	feeds, err := s.Db.GetFeedFollowsByUser(context.Background(), s.Cfg.Username)
	if err != nil {
		return err
	}

	if len(feeds) == 0 {
		fmt.Printf("No feeds found for user %s", s.Cfg.Username)
		return nil
	}

	fmt.Printf("Feeds found for user %s:\n", s.Cfg.Username)

	for _, feed := range feeds {
		fmt.Printf("Feed Name: %s, Feed URL: %s\n", feed.FeedName, feed.FeedUrl)
	}

	return nil
}

func handleUnfollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return errors.New("Wrong number of args for command, Required: 1")
	}

	url := cmd.Args[0]

	feed, err := s.Db.GetFeed(context.Background(), url)
	if err != nil {
		return err
	}

	var params database.DeleteFeedFollowParams
	params.UserID = user.ID
	params.FeedID = feed.ID

	err = s.Db.DeleteFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}

	return nil
}

func handleBrowse(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) > 1 {
		return errors.New("Wrong number of args for command, Required: 0-1")
	}

	var limit int32

	if cmd.Args[0] == "" {
		limit = 2
	} else {
		limit64, err := strconv.ParseInt(cmd.Args[0], 0, 32)

		limit = int32(limit64)

		if err != nil {
			return err
		}
	}

	var params database.GetPostByUserParams
	params.UserID = user.ID
	params.Limit = limit

	posts, err := s.Db.GetPostByUser(context.Background(), params)
	if err != nil {
		return err
	}

	if len(posts) == 0 {
		fmt.Printf("No posts found\n")
	}

	for _, post := range posts {
		fmt.Printf("Title: %s\n", post.Title)
		fmt.Printf("Date: %v\n", post.PublishedAt)
		fmt.Printf("Decs: %s\n", post.Description)
		fmt.Printf("\n")
	}
	return nil
}

func GetCommands() Commands {
	var cmds Commands

	cmds.CommandMap = make(map[string]func(*State, Command) error)

	cmds.register("login", handlerLogin)
	cmds.register("register", handleRegister)
	cmds.register("reset", handleReset)
	cmds.register("users", handleUsers)
	cmds.register("agg", middlewareLoggedIn(handleAgg))
	cmds.register("addfeed", middlewareLoggedIn(handleAddFeed))
	cmds.register("follow", middlewareLoggedIn(handleFollow))
	cmds.register("following", middlewareLoggedIn(handleFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handleUnfollow))
	cmds.register("browse", middlewareLoggedIn(handleBrowse))

	return cmds
}

func GetCommand(name string, args []string) Command {
	var cmd Command
	cmd.Name = name
	cmd.Args = args

	return cmd
}

func middlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(s *State, cmd Command) error {
	return func(s *State, cmd Command) error {
		user, err := s.Db.GetUser(context.Background(), s.Cfg.Username)
		if err != nil {
			return err
		}

		err = handler(s, cmd, user)
		if err != nil {
			return err
		}

		return nil
	}
}
