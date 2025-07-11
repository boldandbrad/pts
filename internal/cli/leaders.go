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

type LeadersOpts struct {
	year                string
	team                string
	minPlateAppearances int16
}

var leadersOpts = LeadersOpts{}
var leadersCmd = &cobra.Command{
	Use:   "leaders [flags]",
	Short: "Show PtS leaderboard for all players.",
	Long:  `Show PtS leaderboard for all players.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		execLeaders()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(leadersCmd)
	leadersCmd.Flags().StringVarP(&leadersOpts.year, "year", "y", fmt.Sprint(time.Now().Year()), "Season to view leaderboard for. Ex: 2025")
	leadersCmd.Flags().StringVarP(&leadersOpts.team, "team", "t", "all", "Team to show leaderboard for.")
	leadersCmd.Flags().Int16VarP(&leadersOpts.minPlateAppearances, "min-plate-appearances", "p", -1, "Only show players with the given number of plate appearances or more. Default: Show all qualified players.")
}

func execLeaders() {
	// check if provided team key is valid
	teamKey := strings.ToUpper(leadersOpts.team)
	if leadersOpts.team != "all" {
		ok := utils.ValidateTeamKey(teamKey)
		if !ok {
			fmt.Printf("Error: invalid team key: %s\n", leadersOpts.team)
			return
		}
	}

	// check if provided year is valid
	// TODO: take into account when teams entered the league
	yearInt := utils.MustBeInt(teamOpts.year)
	currentYear := time.Now().Year()
	if yearInt > time.Now().Year() || yearInt <= 1900 {
		fmt.Printf("Error: year must be between 1901 and %d\n", currentYear)
		return
	}

	battingDataTable, fieldingDataTable, err := utils.FetchTeamStats(leadersOpts.team, leadersOpts.year, leadersOpts.minPlateAppearances)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// parse data into sticks
	sticks := utils.ParseSticks(battingDataTable.Rows, fieldingDataTable.Rows)

	p := tea.NewProgram(tui.NewModel(sticks, leadersOpts.year, teamKey, leadersOpts.minPlateAppearances))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
