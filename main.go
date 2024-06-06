package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// fetchHTML fetches the HTML content from a given URL.
func fetchHTML(url string) (*goquery.Document, error) {
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

func main() {
	url := "https://www.fangraphs.com/leaders-legacy.aspx/major-league?pos=all&stats=bat&lg=all&type=0&season=2024&month=0&season1=2024&ind=0&team=6&qual=1"
	doc, err := fetchHTML(url)
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

	fmt.Printf("Table data successfully written to %s\n", csvFile)
}
