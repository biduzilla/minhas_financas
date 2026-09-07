package datetime

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func ParseMonthYear(monthStr, yearStr string) (time.Time, time.Time, error) {
	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid month")
	}
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid year")
	}
	loc := time.UTC
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)
	end := start.AddDate(0, 1, -1)
	return start, end, nil
}

func ValidateDate(data string) bool {
	pattern := `^(0[1-9]|[12][0-9]|3[01])/(0[1-9]|1[012])/(19|20)\d\d$`
	matched, _ := regexp.MatchString(pattern, data)
	if !matched {
		return false
	}
	parts := strings.Split(data, "/")
	if len(parts) != 3 {
		return false
	}
	day, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])
	year, _ := strconv.Atoi(parts[2])
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return date.Year() == year && int(date.Month()) == month && date.Day() == day
}

func ValidateDateISO(data string) bool {
	pattern := `^(19|20)\d\d-(0[1-9]|1[012])-(0[1-9]|[12][0-9]|3[01])$`
	matched, _ := regexp.MatchString(pattern, data)
	if !matched {
		return false
	}
	parts := strings.Split(data, "-")
	if len(parts) != 3 {
		return false
	}
	year, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])
	day, _ := strconv.Atoi(parts[2])
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return date.Year() == year && int(date.Month()) == month && date.Day() == day
}
