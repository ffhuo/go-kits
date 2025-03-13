package ctime

import "time"

const (
	TimeLayout  = "2006-01-02 15:04:05"
	DateLayout  = "2006-01-02"
	MonthLayout = "2006-01"
)

func Format(t time.Time) string {
	return t.Format("20060102150405")
}

func TimeFormat(t time.Time) string {
	return t.Format(TimeLayout)
}

func DateFormat(t time.Time) string {
	return t.Format(DateLayout)
}

func MonthFormat(t time.Time) string {
	return t.Format(MonthLayout)
}

func ParseTime(str string) (time.Time, error) {
	return time.ParseInLocation(TimeLayout, str, time.Now().Location())
}

func ParseDate(str string) (time.Time, error) {
	return time.ParseInLocation(DateLayout, str, time.Now().Location())
}

func ParseMonth(str string) (time.Time, error) {
	return time.ParseInLocation(MonthLayout, str, time.Now().Location())
}
