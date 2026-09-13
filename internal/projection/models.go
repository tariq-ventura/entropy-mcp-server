package projection

import "time"

type EquipmentType struct {
	Code      string    `json:"code" gorm:"primaryKey;size:100"`
	Name      string    `json:"name" gorm:"size:150;not null"`
	Active    bool      `json:"active" gorm:"not null;default:true"`
	SyncToken string    `json:"-" gorm:"size:36;not null;index"`
	SyncedAt  time.Time `json:"syncedAt" gorm:"not null;index"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (EquipmentType) TableName() string { return "integration_equipment_types" }

type EquipmentTypeAlias struct {
	Key         string    `json:"-" gorm:"primaryKey;size:250"`
	Source      string    `json:"source" gorm:"size:30;not null;index"`
	SourceValue string    `json:"sourceValue" gorm:"size:150;not null"`
	Normalized  string    `json:"normalized" gorm:"size:150;not null;index"`
	TypeCode    string    `json:"typeCode" gorm:"size:100;not null;index"`
	SyncToken   string    `json:"-" gorm:"size:36;not null;index"`
	SyncedAt    time.Time `json:"syncedAt" gorm:"not null;index"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (EquipmentTypeAlias) TableName() string { return "integration_equipment_type_aliases" }

type UnifiedEquipment struct {
	EquipmentKey         string    `json:"equipmentKey" gorm:"primaryKey;size:100"`
	EquipmentTypeCode    string    `json:"equipmentTypeCode" gorm:"size:100;not null;index"`
	Linked               bool      `json:"linked" gorm:"not null;index"`
	Available            bool      `json:"available" gorm:"not null;index"`
	Name                 string    `json:"name" gorm:"size:200"`
	AssetNumber          string    `json:"assetNumber,omitempty" gorm:"size:200;index"`
	Type                 string    `json:"type" gorm:"size:150;index"`
	StartrackType        string    `json:"startrackType,omitempty" gorm:"size:150;index"`
	PrismaStatus         string    `json:"prismaStatus,omitempty" gorm:"size:50;index"`
	TrackingStatus       string    `json:"trackingStatus,omitempty" gorm:"size:100;index"`
	Company              string    `json:"company,omitempty" gorm:"size:150"`
	Year                 int       `json:"year,omitempty"`
	Color                string    `json:"color,omitempty" gorm:"size:100"`
	Brand                string    `json:"brand,omitempty" gorm:"size:150"`
	Model                string    `json:"model,omitempty" gorm:"size:150"`
	GroupName            string    `json:"group,omitempty" gorm:"column:group_name;size:150"`
	Tags                 string    `json:"tags,omitempty" gorm:"type:text"`
	Driver               string    `json:"driver,omitempty" gorm:"size:200"`
	RemoteID             string    `json:"remoteId,omitempty" gorm:"size:150"`
	StartrackDescription string    `json:"startrackDescription,omitempty" gorm:"size:200"`
	PrismaID             string    `json:"prismaId,omitempty" gorm:"size:36;index"`
	StartrackID          string    `json:"startrackId,omitempty" gorm:"size:36;index"`
	Conflicts            []string  `json:"conflicts,omitempty" gorm:"serializer:json;type:text"`
	SyncToken            string    `json:"-" gorm:"size:36;not null;index"`
	SyncedAt             time.Time `json:"syncedAt" gorm:"not null;index"`
}

func (UnifiedEquipment) TableName() string { return "integration_unified_equipments" }

type UnifiedRequest struct {
	ID                string    `json:"id" gorm:"primaryKey;size:36"`
	Project           string    `json:"project" gorm:"size:250;index"`
	Type              string    `json:"type" gorm:"size:150;index"`
	EquipmentTypeCode string    `json:"equipmentTypeCode" gorm:"size:100;index"`
	Requester         string    `json:"requester" gorm:"size:200;index"`
	StartDate         string    `json:"startDate" gorm:"size:40"`
	EndDate           string    `json:"endDate" gorm:"size:40"`
	Status            string    `json:"status" gorm:"size:50;index"`
	Machinery         *string   `json:"machinery,omitempty" gorm:"size:200"`
	EquipmentKey      *string   `json:"equipmentKey,omitempty" gorm:"size:100;index"`
	SourceCreatedAt   string    `json:"sourceCreatedAt,omitempty" gorm:"size:40"`
	SourceUpdatedAt   string    `json:"sourceUpdatedAt,omitempty" gorm:"size:40"`
	SyncToken         string    `json:"-" gorm:"size:36;not null;index"`
	SyncedAt          time.Time `json:"syncedAt" gorm:"not null;index"`
}

func (UnifiedRequest) TableName() string { return "integration_unified_requests" }

type UnifiedMaintenance struct {
	ID              string    `json:"id" gorm:"primaryKey;size:36"`
	EquipmentKey    string    `json:"equipmentKey" gorm:"size:100;index"`
	Vehicle         string    `json:"vehicle" gorm:"size:200;index"`
	Reference       string    `json:"reference" gorm:"type:text"`
	ServiceDate     string    `json:"serviceDate" gorm:"size:40;index"`
	Odometer        float64   `json:"odometer"`
	ServiceTime     string    `json:"serviceTime" gorm:"size:5"`
	HourMeter       float64   `json:"hourMeter"`
	RepairReason    string    `json:"repairReason" gorm:"size:200"`
	Provider        string    `json:"provider,omitempty" gorm:"size:200"`
	Mechanic        string    `json:"mechanic,omitempty" gorm:"size:200"`
	ServiceType     string    `json:"serviceType,omitempty" gorm:"size:200"`
	SourceCreatedAt string    `json:"sourceCreatedAt,omitempty" gorm:"size:40"`
	SourceUpdatedAt string    `json:"sourceUpdatedAt,omitempty" gorm:"size:40"`
	SyncToken       string    `json:"-" gorm:"size:36;not null;index"`
	SyncedAt        time.Time `json:"syncedAt" gorm:"not null;index"`
}

func (UnifiedMaintenance) TableName() string { return "integration_unified_maintenance" }

type UnifiedAssignment struct {
	ID              string    `json:"id" gorm:"primaryKey;size:36"`
	RequestID       string    `json:"requestId" gorm:"size:36;not null;index"`
	EquipmentKey    string    `json:"equipmentKey" gorm:"size:100;index"`
	TaskExternalID  string    `json:"taskExternalId" gorm:"size:100;index"`
	Status          string    `json:"status" gorm:"size:50;not null;index"`
	Title           string    `json:"title" gorm:"size:200"`
	Description     string    `json:"description" gorm:"type:text"`
	TaskType        string    `json:"taskType" gorm:"size:100"`
	ScheduledDate   string    `json:"scheduledDate" gorm:"size:40"`
	Origin          string    `json:"origin" gorm:"size:250"`
	Destination     string    `json:"destination" gorm:"size:250"`
	Latitude        int64     `json:"latitude"`
	Longitude       int64     `json:"longitude"`
	Assignee        string    `json:"assignee" gorm:"size:200"`
	PrismaStatus    string    `json:"prismaStatus" gorm:"size:50"`
	StartrackStatus string    `json:"startrackStatus" gorm:"size:100"`
	Inconsistent    bool      `json:"inconsistent" gorm:"not null;index"`
	SyncToken       string    `json:"-" gorm:"size:36;not null;index"`
	SyncedAt        time.Time `json:"syncedAt" gorm:"not null;index"`
}

func (UnifiedAssignment) TableName() string { return "integration_unified_assignments" }

type SyncConflict struct {
	ID             string    `json:"id" gorm:"primaryKey;size:36"`
	EntityType     string    `json:"entityType" gorm:"size:50;index"`
	EntityKey      string    `json:"entityKey" gorm:"size:150;index"`
	Field          string    `json:"field" gorm:"size:100"`
	PrismaValue    string    `json:"prismaValue,omitempty" gorm:"type:text"`
	StartrackValue string    `json:"startrackValue,omitempty" gorm:"type:text"`
	Message        string    `json:"message" gorm:"type:text"`
	SyncToken      string    `json:"-" gorm:"size:36;not null;index"`
	DetectedAt     time.Time `json:"detectedAt" gorm:"not null;index"`
}

func (SyncConflict) TableName() string { return "integration_sync_conflicts" }

type SyncState struct {
	ID               uint       `json:"-" gorm:"primaryKey"`
	Status           string     `json:"status" gorm:"size:30;not null"`
	LastStartedAt    *time.Time `json:"lastStartedAt,omitempty"`
	LastSucceededAt  *time.Time `json:"lastSucceededAt,omitempty"`
	LastError        string     `json:"lastError,omitempty" gorm:"type:text"`
	EquipmentCount   int64      `json:"equipmentCount"`
	RequestCount     int64      `json:"requestCount"`
	AssignmentCount  int64      `json:"assignmentCount"`
	MaintenanceCount int64      `json:"maintenanceCount"`
	ConflictCount    int64      `json:"conflictCount"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

func (SyncState) TableName() string { return "integration_sync_state" }

type Snapshot struct {
	Token       string
	SyncedAt    time.Time
	Types       []EquipmentType
	Aliases     []EquipmentTypeAlias
	Equipment   []UnifiedEquipment
	Requests    []UnifiedRequest
	Maintenance []UnifiedMaintenance
	Assignments []UnifiedAssignment
	Conflicts   []SyncConflict
}

type Dashboard struct {
	EquipmentTotal     int64      `json:"equipmentTotal"`
	EquipmentAvailable int64      `json:"equipmentAvailable"`
	RequestsPending    int64      `json:"requestsPending"`
	AssignmentsActive  int64      `json:"assignmentsActive"`
	Conflicts          int64      `json:"conflicts"`
	LastSyncedAt       *time.Time `json:"lastSyncedAt,omitempty"`
}

type EquipmentDetail struct {
	Equipment   UnifiedEquipment     `json:"equipment"`
	Maintenance []UnifiedMaintenance `json:"maintenance"`
}
