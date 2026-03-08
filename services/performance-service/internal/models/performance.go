package models

type Performance struct {
	ID            int64   `json:"id"`
	IndicatorName string  `json:"indicator_name"`
	Target        float64 `json:"target"`
	Realization   float64 `json:"realization"`
}
