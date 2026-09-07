package domain

type Recommendation struct {
	EquipmentID               string   `json:"equipmentId"`
	Code                      string   `json:"code"`
	Type                      string   `json:"type"`
	Brand                     string   `json:"brand"`
	Model                     string   `json:"model"`
	SerialNumber              string   `json:"serialNumber"`
	Year                      int      `json:"year"`
	CapacityTons              float64  `json:"capacityTons"`
	Location                  Location `json:"location"`
	DistanceKM                float64  `json:"distanceKm"`
	EngineHours               float64  `json:"engineHours"`
	NextMaintenanceHours      float64  `json:"nextMaintenanceHours"`
	MaintenanceHoursRemaining float64  `json:"maintenanceHoursRemaining"`
	FuelPercent               float64  `json:"fuelPercent"`
	Score                     float64  `json:"score"`
	Reasons                   []string `json:"reasons"`
}

type Recommendations struct {
	RequestID       string           `json:"requestId"`
	Count           int              `json:"count"`
	Recommendations []Recommendation `json:"recommendations"`
}
