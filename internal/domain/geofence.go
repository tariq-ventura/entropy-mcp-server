package domain

type Geofence struct {
	ID               string  `json:"id"`
	GeofenceID       string  `json:"geofenceId"`
	Name             string  `json:"name"`
	Group            string  `json:"group"`
	AdditionalMargin float64 `json:"additionalMargin"`
	Latitude         int64   `json:"latitude"`
	Longitude        int64   `json:"longitude"`
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`
}
type CreateGeofenceRequest struct {
	GeofenceID       string  `json:"geofenceId"`
	Name             string  `json:"name"`
	Group            string  `json:"group,omitempty"`
	AdditionalMargin float64 `json:"additionalMargin"`
	Latitude         int64   `json:"latitude"`
	Longitude        int64   `json:"longitude"`
}
type UnifiedGeofence struct {
	Geofence  Geofence `json:"geofence"`
	Latitude  float64  `json:"latitude"`
	Longitude float64  `json:"longitude"`
}
type UnifiedGeofenceList struct {
	Count     int               `json:"count"`
	Geofences []UnifiedGeofence `json:"geofences"`
}
