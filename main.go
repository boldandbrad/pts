package main

import (
	"github.com/boldandbrad/pts/internal/cli"
)

func main() {

	// TODO: add leaders sub command to show leaderboard. Filters by team(s) and year(s)
	// TODO: add command line flag and env variable for cache directory location
	// TODO: add a player search sub command to search for players
	// TODO: add teams sub command to list available team keys
	// TODO: add option to not include fielding errors
	// TODO: add option to output to CSV

	cli.Execute()
}
