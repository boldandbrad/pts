package cli

import (
	"fmt"
	"log"
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

		pick(args[0])
		return nil
	},
}

var (
	Year string
)

func init() {
	rootCmd.AddCommand(teamCmd)
	teamCmd.Flags().StringVarP(&Year, "season", "s", fmt.Sprint(time.Now().Year()), "Season to view stats for")
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

func pick(teamKey string) {
	// TODO: convert teamKey to uppercase and check that it is a valid team

	// check if year is valid
	// TODO: take into account when teams entered the league
	yearInt := utils.MustBeInt(Year)
	currentYear := time.Now().Year()
	if yearInt > time.Now().Year() || yearInt <= 1900 {
		fmt.Printf("Error: year must be between 1901 and %d\n", currentYear)
		return
	}

	// use cache instead of re-fetching, except always re-fetch the current year
	battingDataTable, fieldingDataTable, err := utils.ReadStatCache(teamKey, Year)
	// TODO: find way to cleanly skip cache check if current year
	if err != nil || yearInt == currentYear {
		// fetch data if it didn't exist in cache
		battingDoc, err := utils.FetchHTML(teamKey, Year, true)
		if err != nil {
			fmt.Printf("Error fetching batting HTML: %v\n", err)
			return
		}
		battingDataTable, err = utils.ExtractTableData(battingDoc)
		if err != nil {
			fmt.Printf("Error extracting batting data from HTML: %v\n", err)
			return
		}

		fieldingDoc, err := utils.FetchHTML(teamKey, Year, false)
		if err != nil {
			fmt.Printf("Error fetching fielding HTML: %v\n", err)
			return
		}
		fieldingDataTable, err = utils.ExtractTableData(fieldingDoc)
		if err != nil {
			fmt.Printf("Error extracting fielding data from HTML: %v\n", err)
			return
		}

		// write fetched data to cache
		// TODO: decide whether to cache the underlying stats or just the table data or both
		utils.WriteStatCache(teamKey, Year, battingDataTable, fieldingDataTable)
	}

	sticks := parseSticks(battingDataTable.Rows, fieldingDataTable.Rows)

	p := tea.NewProgram(tui.NewModel(sticks, Year, teamKey))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
