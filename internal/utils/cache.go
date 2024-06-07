package utils

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

var cacheDir string = "pts_cache"

type DataTable struct {
	Headers []string
	Rows    [][]string
}

func MustBeInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

func WriteStatCache(team string, year string, batDataTable DataTable, fldDataTable DataTable) error {
	// create cache directory if it doesn't exist
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		if err := os.Mkdir(cacheDir, 0755); err != nil {
			fmt.Printf("Error creating cache directory: %v\n", err)
			return err
		}
	}

	var battingData [][]string
	battingData = append(battingData, batDataTable.Headers)
	battingData = append(battingData, batDataTable.Rows...)
	batCSVFile := fmt.Sprintf("%s/%s-%s-bat.csv", cacheDir, team, year)
	if err := WriteCSV(battingData, batCSVFile); err != nil {
		fmt.Printf("Error writing cache file: %v\n", err)
	}

	var fieldingData [][]string
	fieldingData = append(fieldingData, fldDataTable.Headers)
	fieldingData = append(fieldingData, fldDataTable.Rows...)
	fldCSVFile := fmt.Sprintf("%s/%s-%s-fld.csv", cacheDir, team, year)
	if err := WriteCSV(fieldingData, fldCSVFile); err != nil {
		fmt.Printf("Error writing cache file: %v\n", err)
	}
	return nil
}

func ReadStatCache(team string, year string) (DataTable, DataTable, error) {
	if MustBeInt(year) == time.Now().Year() {
		return DataTable{}, DataTable{}, fmt.Errorf("ignoring cache for current year")
	}

	// no cache to read if cache directory doesn't exist
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return DataTable{}, DataTable{}, fmt.Errorf("cache directory does not exist: %v", err)
	}

	batCSVFile := fmt.Sprintf("%s/%s-%s-bat.csv", cacheDir, team, year)
	batData, err := ReadCSV(batCSVFile)
	if err != nil || len(batData) == 0 {
		return DataTable{}, DataTable{}, fmt.Errorf("failed to read cache file: %v", err)
	}
	batDataTable := DataTable{
		Headers: batData[0],
		Rows:    batData[1:],
	}

	fldCSVFile := fmt.Sprintf("%s/%s-%s-fld.csv", cacheDir, team, year)
	fldData, err := ReadCSV(fldCSVFile)
	if err != nil || len(fldData) == 0 {
		return DataTable{}, DataTable{}, fmt.Errorf("failed to read cache file: %v", err)
	}
	fldDataTable := DataTable{
		Headers: fldData[0],
		Rows:    fldData[1:],
	}

	return batDataTable, fldDataTable, nil
}
