package main

import _ "github.com/lib/pq"

import(
	"fmt"
	"os"
	"log"
	"database/sql"

	"github.com/Bryanthai/blog-aggregator/internal/config"
	"github.com/Bryanthai/blog-aggregator/internal/database"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		return
	}

	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)
	var state State
	state.Cfg = &cfg
	state.Db = dbQueries
	commands := Commands{}
	commandinit(&commands)
	args := os.Args
	if len(args) < 2 {
		fmt.Printf("Command name needed!\n")
		os.Exit(1)
	}
	var command Command
	command.Name = os.Args[1]
	command.Arguments = os.Args[2:]
	err = commands.run(&state, command)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func commandinit(commands *Commands) {
	commands.Handlers = make(map[string]func(*State, Command)error)
	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)
	commands.register("users", handlerUsers)
	commands.register("agg", handlerAgg)
	commands.register("addfeed", handlerAddFeed)
	commands.register("feeds", handlerFeeds)
	commands.register("follow", handlerFollow)
	commands.register("following", handlerFollowing)
	commands.register("unfollow", handlerUnfollow)
}