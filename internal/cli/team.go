package cli

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/boldandbrad/pts/internal/pkg/structs"
	"github.com/boldandbrad/pts/internal/tui"
	"github.com/boldandbrad/pts/internal/utils"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var teamCmd = &cobra.Command{
	Use:   "team [flags] teamKey",
	Short: "Show PtS stats for all players on the given team.",
	Long:  `Show PtS stats for all players on the given team.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 || len(args) > 1 {
			return cmd.Help()
		}

		team(args[0])
		return nil
	},
}

var (
	Year string
	// QualifiedOnly bool
)

func init() {
	rootCmd.AddCommand(teamCmd)
	teamCmd.Flags().StringVarP(&Year, "season", "s", fmt.Sprint(time.Now().Year()), "Season to view stats for")
	// teamCmd.Flags().BoolVarP(&QualifiedOnly, "qualified", "q", false, "Only show qualified players")
}

func parseSticks(batRows [][]string, fldRows [][]string) []structs.Stick {
	var sticks []structs.Stick
	for _, row := range batRows {
		stick := structs.Stick{
			Name:                  row[1],
			PlateAppearances:      utils.MustBeInt(row[4]),
			Singles:               utils.MustBeInt(row[6]),
			Doubles:               utils.MustBeInt(row[7]),
			Triples:               utils.MustBeInt(row[8]),
			HomeRuns:              utils.MustBeInt(row[9]),
			RunsScored:            utils.MustBeInt(row[10]),
			RunsBattedIn:          utils.MustBeInt(row[11]),
			StolenBases:           utils.MustBeInt(row[19]),
			Walks:                 utils.MustBeInt(row[12]),
			HitByPitches:          utils.MustBeInt(row[15]),
			SacBunts:              utils.MustBeInt(row[17]),
			StrikeOuts:            utils.MustBeInt(row[14]),
			GroundIntoDoublePlays: utils.MustBeInt(row[18]),
			Errors:                0,
		}
		sticks = append(sticks, stick)
	}

	// parse errors from fielding data
	for idx, stick := range sticks {
		for _, row := range fldRows {
			if row[1] == stick.Name {
				sticks[idx].Errors += utils.MustBeInt(row[8])
			}
		}
	}

	return sticks
}

func team(teamKey string) {
	// check if provided team key is valid
	teamKey = strings.ToUpper(teamKey)
	ok := utils.ValidateTeamKey(teamKey)
	if !ok {
		fmt.Printf("Error: invalid team key: %s\n", teamKey)
		return
	}

	// check if provided year is valid
	// TODO: take into account when teams entered the league
	yearInt := utils.MustBeInt(Year)
	currentYear := time.Now().Year()
	if yearInt > time.Now().Year() || yearInt <= 1900 {
		fmt.Printf("Error: year must be between 1901 and %d\n", currentYear)
		return
	}

	var battingDataTable utils.DataTable
	var fieldingDataTable utils.DataTable
	var err error = nil

	// use cache instead of re-fetching, except always re-fetch the current year
	if yearInt != currentYear {
		battingDataTable, fieldingDataTable, err = utils.ReadTeamCache(teamKey, Year)
	}
	// fetch data if it didn't exist in cache or current year
	if yearInt == currentYear || err != nil {
		battingDataTable, fieldingDataTable, err = utils.FetchTeamStats(teamKey, Year)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// write fetched data to cache
		// TODO: decide whether to cache the underlying stats or just the table data or both
		utils.WriteTeamCache(teamKey, Year, battingDataTable, fieldingDataTable)
	}

	// parse data into sticks
	sticks := parseSticks(battingDataTable.Rows, fieldingDataTable.Rows)

	p := tea.NewProgram(tui.NewModel(sticks, Year, teamKey))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
