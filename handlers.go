package main

import (
	"fmt"
	"net/http"
	"strings"
)

type TimeInterval struct {
	TimeString string
	IsMarked   bool
}

type HourColumn struct {
	HourLabel     string
	TimeIntervals []TimeInterval
}

type DayRow struct {
	DayLabel       string
	Hours          []HourColumn
	ScheduledHours int
}

var minutesMap = [7]string{":00", ":10", ":20", ":30", ":40", ":50", ":00"}

func ScheduleHandler(w http.ResponseWriter, r *http.Request) []DayRow {
	daysOfWeek := []string{"Mon", "Tues", "Wed", "Thurs", "Fri"}
	hoursOfDay := []string{"8 am", "9 am", "10 am", "11 am", "12 pm", "1 pm", "2 pm", "3 pm", "4 pm", "5 pm", "6 pm"}

	var scheduleData []DayRow

	for _, day := range daysOfWeek {
		var hourColumns []HourColumn
		var hourWithoutAmPm string
		var nextHourWithoutAmPm string
		for h, hour := range hoursOfDay[:len(hoursOfDay)-1] {

			parts := strings.Fields(hour)
			nextHourParts := strings.Fields(hoursOfDay[h+1])
			if len(parts) == 0 || len(nextHourParts) == 0 {
				fmt.Println("Validation failed: An hour in Schedule Handler is empty")
				continue
			}
			hourWithoutAmPm = parts[0]
			nextHourWithoutAmPm = nextHourParts[0]
			intervals := make([]TimeInterval, 6)
			var currentHourStr string
			for i := range 6 {
				if i == 5 {
					currentHourStr = hourWithoutAmPm + minutesMap[i] + "-" + nextHourWithoutAmPm + minutesMap[i+1]
				} else {
					currentHourStr = hourWithoutAmPm + minutesMap[i] + "-" + hourWithoutAmPm + minutesMap[i+1]
				}
				intervals[i].TimeString = currentHourStr
			}

			hourColumns = append(hourColumns, HourColumn{
				HourLabel:     hour,
				TimeIntervals: intervals,
			})
		}
		scheduleData = append(scheduleData, DayRow{
			DayLabel: day,
			Hours:    hourColumns,
		})
	}
	return scheduleData
}
