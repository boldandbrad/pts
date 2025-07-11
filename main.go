package main

import (
	"github.com/boldandbrad/pts/internal/cli"
)

func main() {

	// TODO: add pick sub command to show today's lineup for the given team to pick from
	// TODO: add command line flag and env variable for cache directory location
	// TODO: add a player search sub command to search for players
	// TODO: add teams sub command to list available team keys
	// TODO: add option to not include fielding errors
	// TODO: add option to output to CSV

	cli.Execute()
}
