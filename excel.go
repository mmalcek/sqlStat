package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/xuri/excelize/v2"
)

func createExcel(queries []tQuery) error {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	f.SetCellValue("Sheet1", "A1", "Connecting time")
	f.SetCellValue("Sheet1", "A2", stats.connect.String())
	f.SetCellValue("Sheet1", "C1", "Type")
	f.SetCellValue("Sheet1", "D1", "Sheet")

	for i, query := range queries {
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), query.cathegory)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", i+2), query.Name)
		switch query.cathegory {
		case "data":
			if err := addDataSheet(f, query); err != nil {
				fmt.Printf("error adding sheet: %s, error: %s\n", query.Name, err)
				stats.errors = append(stats.errors, fmt.Sprintf("error adding sheet: %s, error: %s", query.Name, err))
				removeSheet(f, query.Name)
			}
		case "duration":
			if err := addDurationsSheet(f, query); err != nil {
				fmt.Printf("error adding sheet: %s, error: %s\n", query.Name, err)
				stats.errors = append(stats.errors, fmt.Sprintf("error adding sheet: %s, error: %s", query.Name, err))
				removeSheet(f, query.Name)
			}
		default:
			return fmt.Errorf("unknown cathegory: %s", query.cathegory)
		}

	}
	if err := addErrorsSheet(f); err != nil {
		return fmt.Errorf("error adding sheet: %w", err)
	}
	file := "results.xlsx"
	if config.OutputDateMark {
		file = fmt.Sprintf("results_%s.xlsx", time.Now().Format("2006-01-02_15-04-05"))
	}
	if config.OutputPath != "" {
		file = filepath.Join(config.OutputPath, file)
	}

	if err := f.SaveAs(file); err != nil {
		return err
	}
	return nil
}

func addDataSheet(f *excelize.File, query tQuery) error {
	fmt.Println("Adding sheet:", query.Name)
	// create sheet
	_, err := f.NewSheet(query.Name)
	if err != nil {
		return err
	}
	// execute query
	results, columns, err := queryData(query) // columns are ordered as requested in query
	if err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}

	// create headers
	colMap := make(map[string]int, 0)
	for i, column := range columns {
		cellName, err := excelize.CoordinatesToCellName(i+1, 1)
		if err != nil {
			return fmt.Errorf("error converting coordinates to cell name: %w", err)
		}
		if err := f.SetCellValue(query.Name, cellName, column); err != nil {
			return fmt.Errorf("error setting cell value: %w", err)
		}
		colMap[column] = i + 1 // helper to keep order of columns
	}
	// create data
	for i, result := range results {
		for j, value := range result {
			cellName, err := excelize.CoordinatesToCellName(colMap[j], i+2)
			if err != nil {
				return fmt.Errorf("error converting coordinates to cell name: %w", err)
			}
			f.SetCellValue(query.Name, cellName, value)
		}
	}
	return nil
}

func addDurationsSheet(f *excelize.File, query tQuery) error {
	fmt.Println("Adding sheet:", query.Name)
	_, err := f.NewSheet(query.Name)
	if err != nil {
		return err
	}
	durations, err := queryDurations(query)
	if err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}
	format := "hh:mm:ss.000"
	xStyle, err := f.NewStyle(&excelize.Style{CustomNumFmt: &format})
	if err != nil {
		return fmt.Errorf("error creating style: %w", err)
	}

	f.SetCellValue(query.Name, "A1", query.Name)
	f.SetCellValue(query.Name, "B1", "Time taken")
	f.SetCellValue(query.Name, "C1", "Time request")
	f.SetCellValue(query.Name, "D1", "Time response")
	f.SetCellStyle(query.Name, "B2", fmt.Sprintf("B%d", len(durations)+1), xStyle)
	f.SetCellStyle(query.Name, "C2", fmt.Sprintf("C%d", len(durations)+1), xStyle)
	f.SetCellStyle(query.Name, "D2", fmt.Sprintf("D%d", len(durations)+1), xStyle)

	for i, duration := range durations {
		f.SetCellValue(query.Name, fmt.Sprintf("A%d", i+2), i)
		f.SetCellValue(query.Name, fmt.Sprintf("B%d", i+2), duration.duration)
		f.SetCellValue(query.Name, fmt.Sprintf("C%d", i+2), duration.timeRequest)
		f.SetCellValue(query.Name, fmt.Sprintf("D%d", i+2), duration.timeResponse)

	}
	f.AddChart(query.Name, "F1", &excelize.Chart{
		Type: excelize.Line,
		Series: []excelize.ChartSeries{
			{
				Name:       query.Name,
				Categories: fmt.Sprintf("'%s'!$A$2:$A$%d", query.Name, len(durations)+1),
				Values:     fmt.Sprintf("'%s'!$B$2:$B$%d", query.Name, len(durations)+1),
			},
		},
		Title: []excelize.RichTextRun{
			{
				Text: query.Name,
			},
		},
	})
	return nil
}

func addErrorsSheet(f *excelize.File) error {
	if len(stats.errors) > 0 {
		fmt.Println("Adding sheet: errors")
		_, err := f.NewSheet("errors")
		if err != nil {
			return err
		}
		f.SetCellValue("errors", "A1", "Errors")
		for i, err := range stats.errors {
			f.SetCellValue("errors", fmt.Sprintf("A%d", i+2), err)
		}
	}
	return nil
}

func removeSheet(f *excelize.File, sheetName string) error {
	if err := f.DeleteSheet(sheetName); err != nil {
		return err
	}
	return nil
}
