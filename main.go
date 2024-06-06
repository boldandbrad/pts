package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

var teams map[string]int = map[string]int{
	"ARI": 15,
	"ATL": 16,
	"BAL": 2,
	"BOS": 3,
	"CHC": 17,
	"CHW": 4,
	"CIN": 18,
	"CLE": 5,
	"COL": 19,
	"DET": 6,
	"HOU": 21,
	"KCR": 7,
	"LAA": 1,
	"LAD": 22,
	"MIA": 20,
	"MIL": 23,
	"MIN": 8,
	"NYM": 25,
	"NYY": 9,
	"OAK": 10,
	"PHI": 26,
	"PIT": 27,
	"SDP": 29,
	"SEA": 11,
	"SFG": 30,
	"STL": 28,
	"TBR": 12,
	"TEX": 13,
	"TOR": 14,
	"WSN": 24,
}

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

type Stick struct {
	Name                  string
	GamesPlayed           int
	PlateAppearances      int
	Singles               int
	Doubles               int
	Triples               int
	HomeRuns              int
	RunsScored            int
	RunsBattedIn          int
	StolenBases           int
	Walks                 int
	HitByPitches          int
	SacBunts              int
	StrikeOuts            int
	GroundIntoDoublePlays int
	Errors                int
}

func (s Stick) PositivePoints() int {
	return s.Singles + (2 * s.Doubles) + (3 * s.Triples) + (4 * s.HomeRuns) + s.RunsScored + s.RunsBattedIn + s.StolenBases + s.Walks + s.HitByPitches + s.SacBunts
}

func (s Stick) NegativePoints() int {
	return s.StrikeOuts + (2 * s.GroundIntoDoublePlays) + s.Errors
}

type Model struct {
	statTable table.Model
}

// fetchHTML fetches the HTML content from a given URL.
func fetchHTML(teamKey string) (*goquery.Document, error) {
	url := fmt.Sprintf("https://www.fangraphs.com/leaders-legacy.aspx/major-league?pos=all&stats=bat&lg=all&type=0&season=2024&month=0&season1=2024&ind=0&team=%d&qual=1", teams[teamKey])
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

// extractTableData extracts the data from the main HTML table.
func extractTableData(doc *goquery.Document) ([]string, [][]string, error) {
	// Extract headers
	var headers []string
	doc.Find(".rgMasterTable thead tr").Last().Find("th").Each(func(i int, s *goquery.Selection) {
		headers = append(headers, s.Text())
	})

	// Extract rows
	var rows [][]string
	doc.Find(".rgMasterTable tbody tr").Each(func(i int, row *goquery.Selection) {
		rowData := make([]string, len(headers))

		row.Find("td").Each(func(j int, cell *goquery.Selection) {
			text := strings.TrimSpace(cell.Text())
			rowData[j] = text
		})

		rows = append(rows, rowData)
	})

	return headers, rows, nil
}

func MustBeInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}

func parseSticks(rows [][]string) []Stick {
	var sticks []Stick
	for _, row := range rows {
		stick := Stick{
			Name:                  row[1],
			PlateAppearances:      MustBeInt(row[4]),
			Singles:               MustBeInt(row[6]),
			Doubles:               MustBeInt(row[7]),
			Triples:               MustBeInt(row[8]),
			HomeRuns:              MustBeInt(row[9]),
			RunsScored:            MustBeInt(row[10]),
			RunsBattedIn:          MustBeInt(row[11]),
			StolenBases:           MustBeInt(row[19]),
			Walks:                 MustBeInt(row[12]),
			HitByPitches:          MustBeInt(row[15]),
			SacBunts:              MustBeInt(row[17]),
			StrikeOuts:            MustBeInt(row[14]),
			GroundIntoDoublePlays: MustBeInt(row[18]),
			Errors:                0,
		}
		sticks = append(sticks, stick)
	}
	return sticks
}

// writeCSV writes the table data to a CSV file.
func writeCSV(data [][]string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, row := range data {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row to CSV: %v", err)
		}
	}
	return nil
}

func makeRow(stick Stick) table.Row {
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

func NewModel(sticks []Stick) Model {
	tableRows := make([]table.Row, len(sticks))
	for i, stick := range sticks {
		tableRows[i] = makeRow(stick)
	}

	return Model{
		statTable: table.New([]table.Column{
			table.NewColumn(columnKeyName, "Name", 16).WithStyle(lipgloss.NewStyle().Align(lipgloss.Left)),
			table.NewColumn(columnKeyPlateAppearances, "PA", 6),
			table.NewColumn(columnKeyPosPoints, "+P", 6),
			table.NewColumn(columnKeyPosPointsPerPA, "+P/PA", 6),
			table.NewColumn(columnKeyNegPoints, "-P", 6),
			table.NewColumn(columnKeyNegPointsPerPA, "-P/PA", 6),
			table.NewColumn(columnKeyPoints, "P", 6),
			table.NewColumn(columnKeyPointsPerPA, "P/PA", 6).WithStyle(lipgloss.NewStyle().Bold(true)),
		}).WithRows(tableRows).BorderRounded().WithBaseStyle(styleBase).WithPageSize(12).SortByDesc(columnKeyPointsPerPA).Focused(true)}
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

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	// selected := m.statTable.HighlightedRow().Data[0].(string)
	view := lipgloss.JoinVertical(
		lipgloss.Left,
		styleSubtle.Render("Press ctrl+c to quit."),
		m.statTable.View(),
	) + "\n"

	return lipgloss.NewStyle().MarginLeft(1).Render(view)
}

func main() {

	teamKey := "DET"
	doc, err := fetchHTML(teamKey)
	if err != nil {
		fmt.Printf("Error fetching HTML: %v\n", err)
		return
	}

	tableHeaders, tableRows, err := extractTableData(doc)
	if err != nil {
		fmt.Printf("Error extracting table data: %v\n", err)
		return
	}

	var tableData [][]string
	tableData = append(tableData, tableHeaders)
	tableData = append(tableData, tableRows...)

	csvFile := "table.csv"
	if err := writeCSV(tableData, csvFile); err != nil {
		fmt.Printf("Error writing CSV file: %v\n", err)
		return
	}

	// fmt.Printf("%s data successfully written to %s\n", teamKey, csvFile)

	sticks := parseSticks(tableRows)

	p := tea.NewProgram(NewModel(sticks))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
