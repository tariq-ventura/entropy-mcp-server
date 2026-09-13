package domain

type Machinery struct {
	ID             string `json:"id"`
	Company        string `json:"company"`
	AssetNumber    string `json:"assetNumber"`
	Name           string `json:"name"`
	EquipmentClass string `json:"equipmentClass"`
	Status         string `json:"status"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}
type Vehicle struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Type        string `json:"type"`
	Year        int    `json:"year"`
	Color       string `json:"color"`
	Brand       string `json:"brand"`
	Model       string `json:"model"`
	Group       string `json:"group"`
	Tags        string `json:"tags"`
	Driver      string `json:"driver"`
	RemoteID    string `json:"remoteId"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}
type UnifiedEquipment struct {
	EquipmentKey         string   `json:"equipmentKey"`
	Linked               bool     `json:"linked"`
	Available            bool     `json:"available"`
	Name                 string   `json:"name"`
	AssetNumber          string   `json:"assetNumber,omitempty"`
	Type                 string   `json:"type"`
	StartrackType        string   `json:"startrackType,omitempty"`
	PrismaStatus         string   `json:"prismaStatus,omitempty"`
	TrackingStatus       string   `json:"trackingStatus,omitempty"`
	Company              string   `json:"company,omitempty"`
	Year                 int      `json:"year,omitempty"`
	Color                string   `json:"color,omitempty"`
	Brand                string   `json:"brand,omitempty"`
	Model                string   `json:"model,omitempty"`
	Group                string   `json:"group,omitempty"`
	Tags                 string   `json:"tags,omitempty"`
	Driver               string   `json:"driver,omitempty"`
	RemoteID             string   `json:"remoteId,omitempty"`
	StartrackDescription string   `json:"startrackDescription,omitempty"`
	PrismaID             string   `json:"prismaId,omitempty"`
	StartrackID          string   `json:"startrackId,omitempty"`
	Conflicts            []string `json:"conflicts,omitempty"`
}
type UnifiedEquipmentList struct {
	Count     int                `json:"count"`
	Equipment []UnifiedEquipment `json:"equipment"`
}
