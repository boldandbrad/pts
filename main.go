package main

import (
	"fmt"
	"log"
	"time"

	"github.com/boldandbrad/pts/internal/pkg/structs"
	"github.com/boldandbrad/pts/internal/utils"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

const (
	columnKeyName             = "name"
	columnKeyPlateAppearances = "pa"
	columnKeyPosPoints        = "pospts"
	columnKeyPosPointsPerPA   = "posptsperpa"
	columnKeyNegPoints        = "negpts"
	columnKeyNegPointsPerPA   = "negptsperpa"
	columnKeyPoints           = "pts"
	columnKeyPointsPerPA      = "ptsperpa"
)

var (
	styleSubtle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888"))

	styleBase = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#eee")).
			BorderForeground(lipgloss.Color("#89b4fa")).
			Align(lipgloss.Right)
)

type Model struct {
	statTable table.Model
	team      string
	year      string
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

func makeRow(stick structs.Stick) table.Row {
	return table.NewRow(table.RowData{
		columnKeyName: stick.Name,
		columnKeyPlateAppearances: func() string {
			if stick.PlateAppearances > 0 {
				return fmt.Sprintf("%d", stick.PlateAppearances)
			}
			return "0"
		}(),
		columnKeyPosPoints: func() string {
			if stick.PositivePoints() > 0 {
				return fmt.Sprintf("%d", stick.PositivePoints())
			}
			return "0"
		}(),
		columnKeyPosPointsPerPA: func() string {
			if stick.PlateAppearances > 0 {
				return fmt.Sprintf("%.3f", float64(stick.PositivePoints())/float64(stick.PlateAppearances))
			}
			return "0"
		}(),
		columnKeyNegPoints: func() string {
			if stick.NegativePoints() > 0 {
				return fmt.Sprintf("-%d", stick.NegativePoints())
			}
			return "0"
		}(),
		columnKeyNegPointsPerPA: func() string {
			if stick.PlateAppearances > 0 {
				return fmt.Sprintf("-%.3f", float64(stick.NegativePoints())/float64(stick.PlateAppearances))
			}
			return "0"
		}(),
		columnKeyPoints: func() string {
			return fmt.Sprintf("%d", stick.PositivePoints()-stick.NegativePoints())
		}(),
		columnKeyPointsPerPA: func() string {
			if stick.PlateAppearances > 0 {
				return fmt.Sprintf("%.3f", float64(stick.PositivePoints()-stick.NegativePoints())/float64(stick.PlateAppearances))
			}
			return "0"
		}(),
	})

}

func NewModel(sticks []structs.Stick, year string, team string) Model {
	tableRows := make([]table.Row, len(sticks))
	for i, stick := range sticks {
		tableRows[i] = makeRow(stick)
	}

	model := Model{
		statTable: table.New([]table.Column{
			table.NewColumn(columnKeyName, "Name", 20).WithStyle(lipgloss.NewStyle().Align(lipgloss.Left)),
			table.NewColumn(columnKeyPlateAppearances, "PA", 6),
			table.NewColumn(columnKeyPosPoints, "+P", 6),
			table.NewColumn(columnKeyPosPointsPerPA, "+P/PA", 6),
			table.NewColumn(columnKeyNegPoints, "-P", 6),
			table.NewColumn(columnKeyNegPointsPerPA, "-P/PA", 6),
			table.NewColumn(columnKeyPoints, "P", 6),
			table.NewColumn(columnKeyPointsPerPA, "P/PA", 6).WithStyle(lipgloss.NewStyle().Bold(true)),
		}).
			WithRows(tableRows).
			BorderRounded().
			WithBaseStyle(styleBase).
			WithPageSize(12).
			SortByDesc(columnKeyPointsPerPA).
			Focused(true),
		team: team,
		year: year,
	}

	model.updateFooter()

	return model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.statTable, cmd = m.statTable.Update(msg)
	cmds = append(cmds, cmd)

	m.updateFooter()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) updateFooter() {
	// highlightedRow := m.statTable.HighlightedRow()

	footerText := fmt.Sprintf(
		"Viewing %s %s - Pg. %d/%d", m.team, m.year,
		m.statTable.CurrentPage(),
		m.statTable.MaxPages(),
	)

	m.statTable = m.statTable.WithStaticFooter(footerText)
}

func (m Model) View() string {
	// selected := m.statTable.HighlightedRow().Data[0].(string)
	view := lipgloss.JoinVertical(
		lipgloss.Left,
		styleSubtle.Render("Press ctrl+c or esc to quit."),
		m.statTable.View(),
	) + "\n"

	return lipgloss.NewStyle().MarginLeft(1).Render(view)
}

func main() {

	// TODO: parse command line arguments for team and year
	// TODO: add leaders sub command to show leaderboard. Filters by team(s) and year(s)
	// TODO: add command line flag for cache directory location
	// TODO: add a player search sub command to search for players
	// TODO: add teams sub command to list available team keys
	// TODO: add option to not include fielding errors
	// TODO: add option to output to CSV

	teamKey := "DET"
	year := "2023"

	// check if year is valid
	// TODO: take into account when teams entered the league
	yearInt := utils.MustBeInt(year)
	currentYear := time.Now().Year()
	if yearInt > time.Now().Year() || yearInt <= 1900 {
		fmt.Printf("Error: year must be between 1901 and %d\n", currentYear)
		return
	}

	// use cache instead of re-fetching, except always re-fetch the current year
	battingDataTable, fieldingDataTable, err := utils.ReadStatCache(teamKey, year)
	// TODO: find way to cleanly skip cache check if current year
	if err != nil || yearInt == currentYear {
		// fetch data if it didn't exist in cache
		battingDoc, err := utils.FetchHTML(teamKey, year, true)
		if err != nil {
			fmt.Printf("Error fetching batting HTML: %v\n", err)
			return
		}
		battingDataTable, err = utils.ExtractTableData(battingDoc)
		if err != nil {
			fmt.Printf("Error extracting batting data from HTML: %v\n", err)
			return
		}

		fieldingDoc, err := utils.FetchHTML(teamKey, year, false)
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
		utils.WriteStatCache(teamKey, year, battingDataTable, fieldingDataTable)
	}

	sticks := parseSticks(battingDataTable.Rows, fieldingDataTable.Rows)

	p := tea.NewProgram(NewModel(sticks, year, teamKey))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
