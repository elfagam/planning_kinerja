package models

type StrategicPlan struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
	Code string `json:"code"`
	Name string `json:"name"`
}
