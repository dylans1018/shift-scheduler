package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
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
var daysOfWeek = []string{"Mon", "Tues", "Wed", "Thurs", "Fri"}
var hoursOfDay = []string{"8 am", "9 am", "10 am", "11 am", "12 pm", "1 pm", "2 pm", "3 pm", "4 pm", "5 pm", "6 pm"}

func ScheduleHandler(w http.ResponseWriter, r *http.Request) []DayRow {

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

func UpdateScheduleHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Grab the hidden input value sent by HTMX
	slotsJSON := r.FormValue("selected_slots")
	if slotsJSON == "" {
		slotsJSON = "[]"
	}

	var intervalsStringFormat []string
	err := json.Unmarshal([]byte(slotsJSON), &intervalsStringFormat)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	dayToHoursMap := backendDailyHoursCounter(intervalsStringFormat)

	//Construct OOB HTMX response
	w.Header().Set("Content-Type", "text/html")

	var responseHTML strings.Builder
	for _, day := range daysOfWeek {
		// We build an OOB element for EVERY day.
		// If a day has 0 hours, it will reset to (0 hrs) on the screen.
		hours := dayToHoursMap[day]
		fmt.Fprintf(&responseHTML,
			`<span id="counter-%s" hx-swap-oob="true">(%.2f hrs)</span>`,
			day, hours,
		)
	}

	w.Write([]byte(responseHTML.String()))
}

func backendDailyHoursCounter(intervals []string) map[string]float64 {
	//PART 1, find how many intervals per day are selected
	numIntervalsForDay := make(map[string]int, len(daysOfWeek))

	for _, day := range daysOfWeek {
		numIntervalsForDay[day] = 0
	}

	for _, intervalString := range intervals {
		intervalParts := strings.Split(intervalString, "||")

		if len(intervalParts) != 2 {
			fmt.Printf("Invalid interval format (number of items): (%q)\n", intervalString)
			continue
		}

		if slices.Contains(daysOfWeek, intervalParts[0]) {
			numIntervalsForDay[intervalParts[0]] += 1
		} else {
			fmt.Printf("Invalid interval format (day of week): (%q)\n", intervalString)
			continue
		}
	}

	//PART 2, convert time intervals to scheduled hours per day
	timeSelectedForDay := make(map[string]float64, len(daysOfWeek))

	for _, day := range daysOfWeek {
		numIntervals := numIntervalsForDay[day]
		timeSelectedForDay[day] = float64(numIntervals) / 6.0
	}

	return timeSelectedForDay
}
