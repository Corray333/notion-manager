package gsheets

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Corray333/notion-manager/internal/notion"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

var (
	ErrNoTime = errors.New("no last synced time found")
)

const (
	TimeLayout = "02/01/2006 15:04:05"
)

func GetLastSyncedTime(srv *sheets.Service, spreadsheetId string) (int64, error) {

	readRange := "Sheet1!X2"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	for _, row := range resp.Values {
		fmt.Println(row)
	}

	if len(resp.Values) == 0 || len(resp.Values[0]) == 0 {
		return 0, ErrNoTime
	}

	lastSynced, err := time.ParseInLocation(TimeLayout, resp.Values[0][0].(string), time.Local)
	if err != nil {
		return 0, err
	}

	return lastSynced.Unix(), nil
}

func SetLastSyncedTime(lastSyncedTimestamp int64, srv *sheets.Service, spreadsheetId string) error {
	writeRange := "Sheet1!X2"

	lastSynced := time.Unix(lastSyncedTimestamp, 0)

	serialized := lastSynced.Format(TimeLayout)

	// Create a ValueRange with the single value
	var vr sheets.ValueRange
	myval := []interface{}{serialized} // Replace "Your Value" with the value you want to insert
	vr.Values = append(vr.Values, myval)

	// Update the cell with the value
	_, err := srv.Spreadsheets.Values.Update(spreadsheetId, writeRange, &vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return err
	}

	return nil
}

func UpdateGoogleSheets() error {
	b, err := os.ReadFile("../secrets/credentials.json")
	if err != nil {
		return err
	}

	config, err := google.ConfigFromJSON(b, sheets.SpreadsheetsScope)
	if err != nil {
		return err
	}
	client := GetClient(config)

	srv, err := sheets.New(client)
	if err != nil {
		return err
	}

	spreadsheetId := "1dStGuMfFU2Vq2V2xgXLyKUq_j3zYBeP15LA0eUQtTAQ"

	lastSynced, err := GetLastSyncedTime(srv, spreadsheetId)
	if err != nil {
		return err
	}

	readRange := "Sheet1!W:W" // H:H - весь первый столбец
	fullTable, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		return err
	}

	writeRange := "Sheet1!A3:W3"
	var vr sheets.ValueRange
	times, err := notion.GetTimes(lastSynced, "", "")
	if err != nil {
		return err
	}

	fmt.Println(len(times), err)
	for _, timeRaw := range times {
		date, _ := time.Parse(notion.TIME_LAYOUT_IN, timeRaw.Properties.WorkDate.Date.Start)
		date2, _ := time.Parse("2006-01-02", timeRaw.Properties.WorkDate.Date.Start)
		if date.Before(date2) {
			date = date2
		}
		lastSyncedRaw, _ := time.Parse(notion.TIME_LAYOUT_IN, timeRaw.CreatedTime)
		lastSynced = lastSyncedRaw.Unix()
		title := ""
		for _, name := range timeRaw.Properties.WhatDid.Title {
			title += name.PlainText
		}

		rawId, err := findRowIndexByID(fullTable, timeRaw.ID)

		if err != nil {
			return err
		}

		myval := []interface{}{
			fmt.Sprintf(`=HYPERLINK("%s"; "%s")`, timeRaw.URL, title),
			timeRaw.Properties.TotalHours.Number,
			date.Format("02/01/2006"),
			func() string {
				if len(timeRaw.Properties.Task.Relation) == 0 {
					return ""
				}
				url := "https://www.notion.so/"
				id := strings.Join(strings.Split(timeRaw.Properties.Task.Relation[0].ID, "-"), "")
				return fmt.Sprintf(`=HYPERLINK("%s"; "%s")`, url+id, timeRaw.Properties.TaskName.Formula.String)
			}(),
			timeRaw.Properties.ProjectName.Formula.String,
		}
		if len(timeRaw.Properties.WhoDid.People) > 0 {
			myval = append(myval, timeRaw.Properties.WhoDid.People[0].Name)
		} else {
			myval = append(myval, "")
		}
		myval = append(myval, []interface{}{
			// TODO:
			timeRaw.Properties.PayableHours.Formula.Number,
			func() string {
				if len(timeRaw.Properties.Task.Relation) == 0 {
					return ""
				}
				return timeRaw.Properties.Task.Relation[0].ID
			}(),
			timeRaw.Properties.Direction.Select.Name,
			timeRaw.Properties.EstimateHours.Formula.String,
			lastSyncedRaw.Format(TimeLayout),
			func() string {
				if timeRaw.Properties.Payment.Checkbox {
					return "TRUE"
				}
				return "FALSE"
			}(),
			func() string {
				if len(timeRaw.Properties.Project.Rollup.Array) == 0 || len(timeRaw.Properties.Project.Rollup.Array[0].Relation) == 0 {
					return ""
				}
				return timeRaw.Properties.Project.Rollup.Array[0].Relation[0].ID
			}(),
			timeRaw.Properties.Month.Formula.String,
			timeRaw.Properties.BH.Formula.Number,
			timeRaw.Properties.SH.Number,
			timeRaw.Properties.DH.Number,
			timeRaw.Properties.BHGS.Formula.String,
			timeRaw.Properties.MonthNumber.Formula.Number,
			timeRaw.Properties.WeekNumber.Formula.Number,
			timeRaw.Properties.DayNumber.Formula.Number,
			timeRaw.Properties.ProjectStatus.Formula.String,
			timeRaw.ID,
		}...)

		if rawId != -1 {
			fmt.Printf("Обновление на %d: %+v", rawId, myval)
			_, err = srv.Spreadsheets.Values.Update(spreadsheetId, fmt.Sprintf("Sheet1!A%d:W%d", rawId, rawId), &sheets.ValueRange{Values: [][]interface{}{myval}}).ValueInputOption("USER_ENTERED").Do()
			if err != nil {
				return err
			}
			continue
		}

		vr.Values = append(vr.Values, myval)

	}

	_, err = srv.Spreadsheets.Values.Append(spreadsheetId, writeRange, &vr).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		return err
	}

	return SetLastSyncedTime(lastSynced, srv, spreadsheetId)

}

func findRowIndexByID(table *sheets.ValueRange, id string) (int, error) {
	// Определяем диапазон, который будем получать (весь лист)

	// Ищем строку с нужным значением
	for i, row := range table.Values {
		if len(row) > 0 && row[0] == id {
			// Возвращаем индекс строки (в Google Sheets строки индексируются с 1)
			fmt.Println("Found: ", i+1)
			return i + 1, nil
		}
	}

	fmt.Println("Not found")
	// Если значение не найдено, возвращаем -1
	return -1, nil
}
