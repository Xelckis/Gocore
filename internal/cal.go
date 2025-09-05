package utils/cal

import (
	"fmt"
	flag "github.com/spf13/pflag"
	"os"
	"strconv"
	"strings"
	"time"
)

func getMonthLines(year int, month time.Month) []string {
	var lines []string

	monthHeader := fmt.Sprintf("%s %d", month.String(), year)
	centeredHeader := fmt.Sprintf("%*s", -10-len(monthHeader)/2, fmt.Sprintf("%*s", 10+len(monthHeader)/2, monthHeader))
	lines = append(lines, centeredHeader)

	lines = append(lines, "Su Mo Tu We Th Fr Sa")

	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	lastDayOfMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local)

	var currentLine strings.Builder
	currentLine.WriteString(strings.Repeat("   ", int(firstOfMonth.Weekday())))

	for day := 1; day <= lastDayOfMonth.Day(); day++ {
		currentLine.WriteString(fmt.Sprintf("%2d ", day))

		if (day+int(firstOfMonth.Weekday()))%7 == 0 {
			lines = append(lines, strings.TrimRight(currentLine.String(), " "))
			currentLine.Reset()
		}
	}

	if currentLine.Len() > 0 {
		lines = append(lines, strings.TrimRight(currentLine.String(), " "))
	}

	return lines
}

func cal(date []string) {
	if len(date) == 1 {
		year, err := strconv.Atoi(date[0])
		if err != nil {
			year = time.Now().Year()
		}

		fmt.Printf("%27s\n", strconv.Itoa(year))

		for row := 0; row < 4; row++ {
			month1 := time.Month(row*3 + 1)
			month2 := time.Month(row*3 + 2)
			month3 := time.Month(row*3 + 3)

			lines1 := getMonthLines(year, month1)
			lines2 := getMonthLines(year, month2)
			lines3 := getMonthLines(year, month3)

			fmt.Printf("%-22s  %-22s  %-22s\n", lines1[0], lines2[0], lines3[0])
			fmt.Printf("%-22s  %-22s  %-22s\n", lines1[1], lines2[1], lines3[1])

			maxLines := len(lines1)
			if len(lines2) > maxLines {
				maxLines = len(lines2)
			}
			if len(lines3) > maxLines {
				maxLines = len(lines3)
			}

			for i := 2; i < maxLines; i++ {
				line1, line2, line3 := "", "", ""
				if i < len(lines1) {
					line1 = lines1[i]
				}
				if i < len(lines2) {
					line2 = lines2[i]
				}
				if i < len(lines3) {
					line3 = lines3[i]
				}
				fmt.Printf("%-22s  %-22s  %-22s\n", line1, line2, line3)
			}
		}
		return
	}

	month := int(time.Now().Month())
	year := time.Now().Year()
	var err error

	if len(date) >= 1 {
		month, err = strconv.Atoi(date[0])
		if err != nil || month < 1 || month > 12 {
			month = int(time.Now().Month())
		}
	}
	if len(date) >= 2 {
		year, err = strconv.Atoi(date[1])
		if err != nil {
			year = time.Now().Year()
		}
	}

	lines := getMonthLines(year, time.Month(month))
	for _, line := range lines {
		fmt.Println(line)
	}
}

func CalFunc() {
	calCmd := flag.NewFlagSet("cal", flag.ExitOnError)
	calCmd.Parse(os.Args[2:])
	cal(calCmd.Args())

}
