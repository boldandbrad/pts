package utils

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/boldandbrad/pts/internal/pkg/structs"

	"github.com/PuerkitoBio/goquery"
)

var Teams map[string]int = map[string]int{
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

var FangraphsUrl = "https://fangraphs.com/leaders/major-league"

func ValidateTeamKey(teamKey string) bool {
	_, ok := Teams[strings.ToUpper(teamKey)]
	return ok
}

// fetchHTML fetches HTML content from fangraphs for a given team and year.
func fetchHTML(team string, year string, minPAs int16, batting bool) (*goquery.Document, error) {
	var statType string
	if batting {
		statType = "bat"
	} else {
		statType = "fld"
	}

	var qualified string
	if minPAs < 0 {
		qualified = "y"
	} else {
		qualified = fmt.Sprintf("%d", minPAs)
	}

	if team == "all" {
		team = "0"
	} else {
		team = fmt.Sprintf("%d", Teams[strings.ToUpper(team)])
	}

	pageitems := "2000000000"

	url := fmt.Sprintf("%s?pos=all&stats=%s&season=%s&season1=%s&team=%s&qual=%s&type=0&pageitems=%s&startdate=&enddate=", FangraphsUrl, statType, year, year, team, qualified, pageitems)
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
	doc.Find(".table-fixed table thead tr").Last().Find("th").Each(func(i int, s *goquery.Selection) {
		headers = append(headers, s.Text())
	})

	// Extract rows
	var rows [][]string
	doc.Find(".table-fixed table tbody tr").Each(func(i int, row *goquery.Selection) {
		rowData := make([]string, len(headers))

		row.Find("td").Each(func(j int, cell *goquery.Selection) {
			text := strings.TrimSpace(cell.Text())
			rowData[j] = text
		})

		rows = append(rows, rowData)
	})

	dataTable := DataTable{
		Headers: headers,
		Rows:    rows,
	}

	return dataTable, nil
}

func FetchTeamStats(teamKey string, year string, minPAs int16) (DataTable, DataTable, error) {
	battingDoc, err := fetchHTML(teamKey, year, minPAs, true)
	if err != nil {
		return DataTable{}, DataTable{}, fmt.Errorf("error fetching batting HTML: %v", err)
	}
	battingDataTable, err := extractTableData(battingDoc)
	if err != nil {
		return DataTable{}, DataTable{}, fmt.Errorf("error extracting batting data from HTML: %v", err)
	}

	fieldingDoc, err := fetchHTML(teamKey, year, minPAs, false)
	if err != nil {
		return DataTable{}, DataTable{}, fmt.Errorf("error fetching fielding HTML: %v", err)
	}
	fieldingDataTable, err := extractTableData(fieldingDoc)
	if err != nil {
		return DataTable{}, DataTable{}, fmt.Errorf("error extracting fielding data from HTML: %v", err)
	}

	return battingDataTable, fieldingDataTable, nil
}

func ParseSticks(batRows [][]string, fldRows [][]string) []structs.Stick {
	var sticks []structs.Stick
	for _, row := range batRows {
		stick := structs.Stick{
			Name:                  row[1],
			Team:                  row[2],
			PlateAppearances:      MustBeInt(row[5]),
			Singles:               MustBeInt(row[7]),
			Doubles:               MustBeInt(row[8]),
			Triples:               MustBeInt(row[9]),
			HomeRuns:              MustBeInt(row[10]),
			RunsScored:            MustBeInt(row[11]),
			RunsBattedIn:          MustBeInt(row[12]),
			StolenBases:           MustBeInt(row[20]),
			Walks:                 MustBeInt(row[13]),
			HitByPitches:          MustBeInt(row[16]),
			SacBunts:              MustBeInt(row[18]),
			StrikeOuts:            MustBeInt(row[15]),
			GroundIntoDoublePlays: MustBeInt(row[19]),
			Errors:                0,
		}
		if stick.PlateAppearances > 0 {
			sticks = append(sticks, stick)
		}
	}

	// parse errors from fielding data
	for idx, stick := range sticks {
		for _, row := range fldRows {
			if row[1] == stick.Name {
				sticks[idx].Errors += MustBeInt(row[10])
			}
		}
	}

	return sticks
}
