package main

import(
	"os"
	"time"
	"fmt"
	"context"
	"strconv"
	"html"

	"github.com/google/uuid"
	"github.com/Bryanthai/blog-aggregator/internal/config"
	"github.com/Bryanthai/blog-aggregator/internal/database"
)

type State struct{
	Db *database.Queries
	Cfg *config.Config
}

type Command struct{
	Name string
	Arguments []string
}

type Commands struct{
	Handlers map[string]func(*State, Command)error
}

func handlerLogin(s *State, cmd Command) error {
	if len(cmd.Arguments) != 1 {
		fmt.Printf("Incorrect usage of command!\n")
		os.Exit(1)
	}
	dbQueries := s.Db
	user, _ := dbQueries.GetUser(context.Background(), cmd.Arguments[0])
	if (user == database.User{}) {
		fmt.Printf("Username doesn't exists!\n")
		os.Exit(1)
	}
	err := s.Cfg.SetUser(cmd.Arguments[0])
	if err != nil {
		return err
	}
	fmt.Printf("Currently logged in as %s\n", cmd.Arguments[0])
	return nil
}

func handlerReset(s *State, cmd Command) error {
	dbQueries := s.Db
	err := dbQueries.Reset(context.Background())
	fmt.Printf("Database resetted!\n")
	return err
}

func handlerUsers(s *State, cmd Command) error {
	dbQueries := s.Db
	users, err := dbQueries.GetUsers(context.Background())
	currentUser := s.Cfg.Username
	for _, name := range users {
		fmt.Printf("* %s", name)
		if currentUser == name {
			fmt.Printf(" (current)")
		}
		fmt.Printf("\n")
	}
	return err
}

func handlerAgg(s *State, cmd Command) error {
	args := cmd.Arguments
	if len(args) != 1 {
		fmt.Printf("Incorrect usage of command!\n")
		os.Exit(1)
	}

	var interval time.Duration
	interval, err := time.ParseDuration(args[0])
	if err != nil {
		return err
	}
	fmt.Printf("Collecting feeds every %s\n", args[0])

	ticker := time.NewTicker(interval)
	for ; ;<- ticker.C {
		scrapeFeeds(s)
	}
	return nil
}

func handlerAddFeed(s *State, cmd Command) error {
	args := cmd.Arguments
	dbQueries := s.Db
	if len(args) != 2 {
		fmt.Printf("Incorrect usage of command!\n")
		os.Exit(1)
	}

	currentUser, err := dbQueries.GetUser(context.Background(), s.Cfg.Username)
	if err != nil {
		return err
	}

	var feedParams database.CreateFeedParams
	feedParams.ID = uuid.New()
	feedParams.CreatedAt = time.Now()
	feedParams.UpdatedAt = time.Now()
	feedParams.Name = args[0]
	feedParams.Url = args[1]
	feedParams.UserID = currentUser.ID

	feed, err := dbQueries.CreateFeed(context.Background(), feedParams)
	if err != nil {
		return err
	}
	fmt.Printf("New Feed of %s created\n", feed.Url)

	feedid, err := dbQueries.GetFeedName(context.Background(), args[1])
	if err != nil {
		return err
	}

	var followParam database.CreateFeedFollowParams
	followParam.ID = uuid.New()
	followParam.CreatedAt = time.Now()
	followParam.UpdatedAt = time.Now()
	followParam.UserID = currentUser.ID
	followParam.FeedID = feedid

	followFeed, err := dbQueries.CreateFeedFollow(context.Background(), followParam)
	if err != nil {
		return err
	}

	fmt.Printf("User %s has followed %s!\n", followFeed.UserName, followFeed.FeedName)

	return nil
}

func handlerFeeds(s *State, cmd Command) error {
	dbQueries := s.Db

	feeds, err := dbQueries.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	if len(feeds) == 0{
		return fmt.Errorf("No feeds yet!")
	}

	fmt.Printf("Feeds:\n")
	for _, feed := range feeds {
		username, err := dbQueries.GetUsername(context.Background(), feed.UserID)
		if err != nil {
			return err
		}
		fmt.Printf("- Name: %s, Url: %s, Created by: %s\n", feed.Name, feed.Url, username)
	}
	return nil
}

func handlerFollow(s *State, cmd Command) error {
	dbQueries := s.Db
	args := cmd.Arguments
	if len(args) != 1 {
		fmt.Printf("Incorrect usage of command!\n")
		os.Exit(1)
	}

	feedid, err := dbQueries.GetFeedName(context.Background(), args[0])
	if err != nil {
		return err
	}
	user, err := dbQueries.GetUser(context.Background(), s.Cfg.Username)
	if err != nil {
		return err
	}

	var followParam database.CreateFeedFollowParams
	followParam.ID = uuid.New()
	followParam.CreatedAt = time.Now()
	followParam.UpdatedAt = time.Now()
	followParam.UserID = user.ID
	followParam.FeedID = feedid

	followFeed, err := dbQueries.CreateFeedFollow(context.Background(), followParam)
	if err != nil {
		return err
	}

	fmt.Printf("User %s has followed %s!\n", followFeed.UserName, followFeed.FeedName)

	return nil
}

func handlerUnfollow(s *State, cmd Command) error {
	dbQueries := s.Db
	args := cmd.Arguments
	if len(args) != 1 {
		fmt.Printf("Incorrect usage of command!\n")
		os.Exit(1)
	}

	feedid, err := dbQueries.GetFeedName(context.Background(), args[0])
	if err != nil {
		return err
	}
	user, err := dbQueries.GetUser(context.Background(), s.Cfg.Username)
	if err != nil {
		return err
	}

	var unfollowParam database.UnfollowFeedParams
	unfollowParam.UserID = user.ID
	unfollowParam.FeedID = feedid

	err = dbQueries.UnfollowFeed(context.Background(), unfollowParam)
	if err != nil {
		return err
	}

	fmt.Printf("User %s has unfollowed %s!\n", s.Cfg.Username, args[0])

	return nil
}

func handlerBrowse(s *State, cmd Command) error {
	dbQueries := s.Db
	args := cmd.Arguments
	user, err := dbQueries.GetUser(context.Background(), s.Cfg.Username)
	if err != nil {
		return err
	}
	if len(args) > 2 || len(args) < 1{
		fmt.Printf("Incorrect usage of command!\n")
		os.Exit(1)
	}
	feedID, err := dbQueries.GetFeedName(context.Background(), args[0])
	if err != nil {
		return err
	}
	var postsByUser database.GetPostsByUserParams
	if len(args) == 2 {
		limiting , _ := strconv.Atoi(args[1])
		postsByUser.Limit = int32(limiting)
		postsByUser.UserID = user.ID
		postsByUser.FeedID = feedID
		posts, err := dbQueries.GetPostsByUser(context.Background(), postsByUser)
		if err != nil {
			return err
		}

		for i, rPost := range posts {
			fmt.Printf("%d. Title: %v\n", i,rPost.Title)
			fmt.Printf("From URL: %v\n", rPost.Url)
			fmt.Printf("Published at: %v\n", rPost.PublishedAt)
			fmt.Printf("%v\n\n",rPost.Description)
		}
	} else {
		postsByUser.Limit = 2
		postsByUser.UserID = user.ID
		postsByUser.FeedID = feedID
		posts, err := dbQueries.GetPostsByUser(context.Background(), postsByUser)

		if err != nil {
			return err
		}

		for i, rPost := range posts {
			fmt.Printf("%d. Title: %v\n", i,rPost.Title)
			fmt.Printf("From URL: %v\n", rPost.Url)
			fmt.Printf("Published at: %v\n", rPost.PublishedAt)
			fmt.Printf("%v\n\n", rPost.Description)
		}	
	}

	return nil
}

func handlerFollowing(s *State, cmd Command) error {
	dbQueries := s.Db
	followingFeeds, err := dbQueries.GetFeedFollowsForUser(context.Background(), s.Cfg.Username)
	if err != nil {
		return err
	}

	fmt.Printf("You are currently following:\n")
	for _, feeds := range followingFeeds{
		fmt.Printf("- %s\n", feeds.FeedName)
	}
	return nil
}

func handlerRegister(s *State, cmd Command) error {
	args := cmd.Arguments
	dbQueries := s.Db
	if len(args) != 1 {
		fmt.Printf("Incorrect usage of command!\n")
		os.Exit(1)
	}
	var err error
	user, _ := dbQueries.GetUser(context.Background(), args[0])
	if (user != database.User{}) {
		fmt.Printf("Username already exists!\n")
		os.Exit(1)
	}

	var userParam database.CreateUserParams
	userParam.ID = uuid.New()
	userParam.CreatedAt = time.Now()
	userParam.UpdatedAt = time.Now()
	userParam.Name = args[0]

	user, err = dbQueries.CreateUser(context.Background(), userParam)
	if err != nil {
		return err
	}
	fmt.Printf("User %s created\n", user.Name)

	err = s.Cfg.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("Currently logged in as %s\n", user.Name)
	return nil
}

func (c *Commands) run(s *State, cmd Command) error {
	handler, ok := c.Handlers[cmd.Name]
	if !ok {
		return fmt.Errorf("Command doesnt exist!\n")
	}
	err := handler(s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (c *Commands) register(name string, f func(*State, Command) error) {
	c.Handlers[name] = f
	return
}

func scrapeFeeds(s *State) error {
	dbQueries := s.Db
	nextURL, err := dbQueries.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	var markingParam database.MarkFeedFetchedParams
	markingParam.Url = nextURL
	markingParam.LastFetchAt.Time = time.Now()
	markingParam.LastFetchAt.Valid = true
	err = dbQueries.MarkFeedFetched(context.Background(), markingParam)
	if err != nil {
		return err
	}

	Feed, err := fetchFeed(context.Background(), nextURL)
	if err != nil{
		return err
	}
	var postParams database.CreatePostParams
	var upostParams database.UpdatePostParams
	for _, post := range Feed.Channel.Item {
		inPost, _ := dbQueries.CheckPostByURL(context.Background(), post.Link)
		if (inPost != database.Post{}) {
			upostParams.UpdatedAt = time.Now()
			upostParams.Title = post.Title
			upostParams.Description.String = html.UnescapeString(post.Description)
			upostParams.Description.Valid = true
			upostParams.PublishedAt = post.PubDate
			upostParams.Url = post.Link
			err = dbQueries.UpdatePost(context.Background(), upostParams)
			if err != nil {
				return err
			}
		} else {
			postParams.ID = uuid.New()
			postParams.FeedID, _ = dbQueries.GetFeedName(context.Background(), nextURL)
			postParams.CreatedAt = time.Now()
			postParams.UpdatedAt = time.Now()
			postParams.Title = post.Title
			postParams.Url = post.Link
			postParams.Description.String = html.UnescapeString(post.Description)
			postParams.Description.Valid = true
			postParams.PublishedAt = post.PubDate
			_, err = dbQueries.CreatePost(context.Background(), postParams)
			if err != nil {
				return err
			}
		}
	}
	return nil
}