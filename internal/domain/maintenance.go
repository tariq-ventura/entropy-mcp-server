package domain

type Maintenance struct {
	ID           string  `json:"id"`
	Vehicle      string  `json:"vehicle"`
	Reference    string  `json:"reference"`
	ServiceDate  string  `json:"serviceDate"`
	Odometer     float64 `json:"odometer"`
	ServiceTime  string  `json:"serviceTime"`
	HourMeter    float64 `json:"hourMeter"`
	RepairReason string  `json:"repairReason"`
	Provider     string  `json:"provider,omitempty"`
	Mechanic     string  `json:"mechanic,omitempty"`
	ServiceType  string  `json:"serviceType,omitempty"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

type MaintenanceSummary struct {
	RecordCount           int      `json:"recordCount"`
	LatestServiceDate     string   `json:"latestServiceDate,omitempty"`
	LatestServiceType     string   `json:"latestServiceType,omitempty"`
	LatestRepairReason    string   `json:"latestRepairReason,omitempty"`
	LatestReference       string   `json:"latestReference,omitempty"`
	LatestOdometer        float64  `json:"latestOdometer,omitempty"`
	LatestHourMeter       float64  `json:"latestHourMeter,omitempty"`
	RecentCorrectiveCount int      `json:"recentCorrectiveCount"`
	Signals               []string `json:"signals,omitempty"`
}
