package utils

import (
	"strconv"
	"time"
)

func ParseDate(val string) (time.Time, error) {
	// Example parses YYYYMM or YYYYMMDD, adjust to match data
	if len(val) == 6 {
		return time.Parse("200601", val)
	}
	if len(val) == 8 {
		return time.Parse("20060102", val)
	}
	return time.Time{}, nil
}

func ParseInt(val string) int {
	i, _ := strconv.Atoi(val)
	return i
}

func ParseFloat(val string) float64 {
	f, _ := strconv.ParseFloat(val, 64)
	return f
}
