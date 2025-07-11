package cli

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/boldandbrad/pts/internal/tui"
	"github.com/boldandbrad/pts/internal/utils"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type TeamOpts struct {
	year string
}

var teamOpts = TeamOpts{}
var teamCmd = &cobra.Command{
	Use:   "team [flags] teamKey",
	Short: "Show PtS stats for all players on the given team.",
	Long:  `Show PtS stats for all players on the given team.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		execTeam(args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(teamCmd)
	teamCmd.Flags().StringVarP(&teamOpts.year, "year", "y", fmt.Sprint(time.Now().Year()), "Season to view stats for. Ex: 2025")
}

func execTeam(teamKey string) {
	// check if provided team key is valid
	teamKey = strings.ToUpper(teamKey)
	ok := utils.ValidateTeamKey(teamKey)
	if !ok {
		fmt.Printf("Error: invalid team key: %s\n", teamKey)
		return
	}

	// check if provided year is valid
	// TODO: take into account when teams entered the league
	yearInt := utils.MustBeInt(teamOpts.year)
	currentYear := time.Now().Year()
	if yearInt > time.Now().Year() || yearInt <= 1900 {
		fmt.Printf("Error: year must be between 1901 and %d\n", currentYear)
		return
	}

	var minPAs int16 = 0

	var battingDataTable utils.DataTable
	var fieldingDataTable utils.DataTable
	var err error = nil

	// use cache instead of re-fetching, except always re-fetch the current year
	if yearInt != currentYear {
		battingDataTable, fieldingDataTable, err = utils.ReadTeamCache(teamKey, teamOpts.year)
	}
	// fetch data if it didn't exist in cache or current year
	if yearInt == currentYear || err != nil {
		battingDataTable, fieldingDataTable, err = utils.FetchTeamStats(teamKey, teamOpts.year, minPAs)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// write fetched data to cache
		// TODO: decide whether to cache the underlying stats or just the table data or both
		utils.WriteTeamCache(teamKey, teamOpts.year, battingDataTable, fieldingDataTable)
	}

	// parse data into sticks
	sticks := utils.ParseSticks(battingDataTable.Rows, fieldingDataTable.Rows)

	p := tea.NewProgram(tui.NewModel(sticks, teamOpts.year, teamKey, minPAs))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
