package dto

type Processed struct {
	InHour  int `json:"in_hour"`
	InDay   int `json:"in_day"`
	InWeek  int `json:"in_week"`
	InMonth int `json:"in_month"`
}
