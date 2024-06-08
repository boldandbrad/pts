package utils

import (
	"fmt"
	"net/http"
	"strings"

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

func ValidateTeamKey(teamKey string) bool {
	_, ok := Teams[strings.ToUpper(teamKey)]
	return ok
}

// fetchHTML fetches HTML content from fangraphs for a given team and year.
func fetchHTML(teamKey string, year string, batting bool) (*goquery.Document, error) {
	var statType string
	if batting {
		statType = "bat"
	} else {
		statType = "fld"
	}

	url := fmt.Sprintf("https://www.fangraphs.com/leaders-legacy.aspx/major-league?pos=all&stats=%s&lg=all&type=0&season=%s&month=0&season1=%s&ind=0&team=%d&qual=1&page=1_100", statType, year, year, Teams[strings.ToUpper(teamKey)])
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
		Headers: headers,
		Rows:    rows,
	}

	return dataTable, nil
}

func FetchTeamStats(teamKey string, year string) (DataTable, DataTable, error) {
	battingDoc, err := fetchHTML(teamKey, year, true)
	if err != nil {
		return DataTable{}, DataTable{}, fmt.Errorf("error fetching batting HTML: %v", err)
	}
	battingDataTable, err := extractTableData(battingDoc)
	if err != nil {
		return DataTable{}, DataTable{}, fmt.Errorf("error extracting batting data from HTML: %v", err)
	}

	fieldingDoc, err := fetchHTML(teamKey, year, false)
	if err != nil {
		return DataTable{}, DataTable{}, fmt.Errorf("error fetching fielding HTML: %v", err)
	}
	fieldingDataTable, err := extractTableData(fieldingDoc)
	if err != nil {
		return DataTable{}, DataTable{}, fmt.Errorf("error extracting fielding data from HTML: %v", err)
	}

	return battingDataTable, fieldingDataTable, nil
}
