package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

var cacheDir string = "cache"

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

type DataTable struct {
	headers []string
	rows    [][]string
}

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
	team      string
	year      string
}

// fetchHTML fetches the HTML content from a given URL.
func fetchHTML(teamKey string, year string, batting bool) (*goquery.Document, error) {
	var statType string
	if batting {
		statType = "bat"
	} else {
		statType = "fld"
	}

	url := fmt.Sprintf("https://www.fangraphs.com/leaders-legacy.aspx/major-league?pos=all&stats=%s&lg=all&type=0&season=%s&month=0&season1=%s&ind=0&team=%d&qual=1&page=1_100", statType, year, year, teams[teamKey])
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
func extractTableData(doc *goquery.Document) (DataTable, error) {
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

	dataTable := DataTable{
		headers: headers,
		rows:    rows,
	}

	return dataTable, nil
}

func MustBeInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

func parseSticks(batRows [][]string, fldRows [][]string) []Stick {
	var sticks []Stick
	for _, row := range batRows {
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

	// parse errors from fielding data
	for idx, stick := range sticks {
		for _, row := range fldRows {
			if row[1] == stick.Name {
				sticks[idx].Errors += MustBeInt(row[8])
			}
		}
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

func readCSV(filename string) ([][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	data, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV data: %v", err)
	}
	return data, nil
}

func writeStatCache(team string, year string, batDataTable DataTable, fldDataTable DataTable) error {
	// create cache directory if it doesn't exist
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		if err := os.Mkdir(cacheDir, 0755); err != nil {
			fmt.Printf("Error creating cache directory: %v\n", err)
			return err
		}
	}

	var battingData [][]string
	battingData = append(battingData, batDataTable.headers)
	battingData = append(battingData, batDataTable.rows...)
	batCSVFile := fmt.Sprintf("%s/%s-%s-bat.csv", cacheDir, team, year)
	if err := writeCSV(battingData, batCSVFile); err != nil {
		fmt.Printf("Error writing cache file: %v\n", err)
	}

	var fieldingData [][]string
	fieldingData = append(fieldingData, fldDataTable.headers)
	fieldingData = append(fieldingData, fldDataTable.rows...)
	fldCSVFile := fmt.Sprintf("%s/%s-%s-fld.csv", cacheDir, team, year)
	if err := writeCSV(fieldingData, fldCSVFile); err != nil {
		fmt.Printf("Error writing cache file: %v\n", err)
	}
	return nil
}

func readStatCache(team string, year string) (DataTable, DataTable, error) {
	if MustBeInt(year) == time.Now().Year() {
		return DataTable{}, DataTable{}, fmt.Errorf("ignoring cache for current year")
	}

	// no cache to read if cache directory doesn't exist
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return DataTable{}, DataTable{}, fmt.Errorf("cache directory does not exist: %v", err)
	}

	batCSVFile := fmt.Sprintf("%s/%s-%s-bat.csv", cacheDir, team, year)
	batData, err := readCSV(batCSVFile)
	if err != nil || len(batData) == 0 {
		return DataTable{}, DataTable{}, fmt.Errorf("failed to read cache file: %v", err)
	}
	batDataTable := DataTable{
		headers: batData[0],
		rows:    batData[1:],
	}

	fldCSVFile := fmt.Sprintf("%s/%s-%s-fld.csv", cacheDir, team, year)
	fldData, err := readCSV(fldCSVFile)
	if err != nil || len(fldData) == 0 {
		return DataTable{}, DataTable{}, fmt.Errorf("failed to read cache file: %v", err)
	}
	fldDataTable := DataTable{
		headers: fldData[0],
		rows:    fldData[1:],
	}

	return batDataTable, fldDataTable, nil
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

func NewModel(sticks []Stick, year string, team string) Model {
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
	yearInt := MustBeInt(year)
	currentYear := time.Now().Year()
	if yearInt > time.Now().Year() || yearInt <= 1900 {
		fmt.Printf("Error: year must be between 1901 and %d\n", currentYear)
		return
	}

	// use cache instead of re-fetching, except always re-fetch the current year
	battingDataTable, fieldingDataTable, err := readStatCache(teamKey, year)
	// TODO: find way to cleanly skip cache check if current year
	if err != nil || yearInt == currentYear {
		// fetch data if it didn't exist in cache
		battingDoc, err := fetchHTML(teamKey, year, true)
		if err != nil {
			fmt.Printf("Error fetching batting HTML: %v\n", err)
			return
		}
		battingDataTable, err = extractTableData(battingDoc)
		if err != nil {
			fmt.Printf("Error extracting batting data from HTML: %v\n", err)
			return
		}

		fieldingDoc, err := fetchHTML(teamKey, year, false)
		if err != nil {
			fmt.Printf("Error fetching fielding HTML: %v\n", err)
			return
		}
		fieldingDataTable, err = extractTableData(fieldingDoc)
		if err != nil {
			fmt.Printf("Error extracting fielding data from HTML: %v\n", err)
			return
		}

		// write fetched data to cache
		// TODO: decide whether to cache the underlying stats or just the table data or both
		writeStatCache(teamKey, year, battingDataTable, fieldingDataTable)
	}

	sticks := parseSticks(battingDataTable.rows, fieldingDataTable.rows)

	p := tea.NewProgram(NewModel(sticks, year, teamKey))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
