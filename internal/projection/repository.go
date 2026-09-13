package projection

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrNotFound = errors.New("projection record not found")

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) (*Repository, error) {
	if db == nil {
		return nil, errors.New("MCP PostgreSQL connection is required")
	}
	return &Repository{db: db}, nil
}

func (r *Repository) Migrate(ctx context.Context) error {
	return r.db.WithContext(ctx).AutoMigrate(&EquipmentType{}, &EquipmentTypeAlias{}, &UnifiedEquipment{}, &UnifiedRequest{}, &UnifiedMaintenance{}, &UnifiedAssignment{}, &SyncConflict{}, &SyncState{})
}

func (r *Repository) BeginSync(ctx context.Context, startedAt time.Time) error {
	state := SyncState{ID: 1, Status: "RUNNING", LastStartedAt: &startedAt, LastError: ""}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoUpdates: clause.AssignmentColumns([]string{"status", "last_started_at", "last_error", "updated_at"})}).Create(&state).Error
}

func (r *Repository) FailSync(ctx context.Context, syncError error) error {
	return r.db.WithContext(ctx).Model(&SyncState{}).Where("id = ?", 1).Updates(map[string]any{"status": "FAILED", "last_error": syncError.Error(), "updated_at": time.Now().UTC()}).Error
}

func (r *Repository) ApplySnapshot(ctx context.Context, snapshot Snapshot) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := upsert(tx, snapshot.Types); err != nil {
			return err
		}
		if err := upsert(tx, snapshot.Aliases); err != nil {
			return err
		}
		if err := upsert(tx, snapshot.Equipment); err != nil {
			return err
		}
		if err := upsert(tx, snapshot.Requests); err != nil {
			return err
		}
		if err := upsert(tx, snapshot.Maintenance); err != nil {
			return err
		}
		if err := upsert(tx, snapshot.Assignments); err != nil {
			return err
		}
		if err := upsert(tx, snapshot.Conflicts); err != nil {
			return err
		}
		for _, model := range []any{&EquipmentType{}, &EquipmentTypeAlias{}, &UnifiedEquipment{}, &UnifiedRequest{}, &UnifiedMaintenance{}, &UnifiedAssignment{}, &SyncConflict{}} {
			if err := tx.Where("sync_token IS NULL OR sync_token <> ?", snapshot.Token).Delete(model).Error; err != nil {
				return err
			}
		}
		completedAt := snapshot.SyncedAt
		state := SyncState{ID: 1, Status: "READY", LastStartedAt: &completedAt, LastSucceededAt: &completedAt, LastError: "", EquipmentCount: int64(len(snapshot.Equipment)), RequestCount: int64(len(snapshot.Requests)), AssignmentCount: int64(len(snapshot.Assignments)), MaintenanceCount: int64(len(snapshot.Maintenance)), ConflictCount: int64(len(snapshot.Conflicts))}
		return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, UpdateAll: true}).Create(&state).Error
	})
}

func upsert[T any](tx *gorm.DB, records []T) error {
	if len(records) == 0 {
		return nil
	}
	return tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&records).Error
}

func (r *Repository) ListAllEquipment(ctx context.Context) ([]UnifiedEquipment, error) {
	items := make([]UnifiedEquipment, 0)
	err := r.db.WithContext(ctx).Order("equipment_key ASC").Find(&items).Error
	return items, err
}

func (r *Repository) ListEquipment(ctx context.Context, page, pageSize int, equipmentType, status, search string, onlyAvailable, onlyLinked bool) ([]UnifiedEquipment, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	query := r.db.WithContext(ctx).Model(&UnifiedEquipment{})
	if equipmentType != "" {
		query = query.Where("LOWER(type) = LOWER(?) OR equipment_type_code = ?", strings.TrimSpace(equipmentType), strings.ToUpper(strings.TrimSpace(equipmentType)))
	}
	if status != "" {
		query = query.Where("LOWER(prisma_status) = LOWER(?) OR LOWER(tracking_status) = LOWER(?)", strings.TrimSpace(status), strings.TrimSpace(status))
	}
	if onlyAvailable {
		query = query.Where("available = ?", true)
	}
	if onlyLinked {
		query = query.Where("linked = ?", true)
	}
	if search != "" {
		pattern := "%" + strings.TrimSpace(search) + "%"
		query = query.Where("equipment_key ILIKE ? OR asset_number ILIKE ? OR name ILIKE ? OR brand ILIKE ? OR model ILIKE ? OR driver ILIKE ?", pattern, pattern, pattern, pattern, pattern, pattern)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := make([]UnifiedEquipment, 0)
	err := query.Order("available DESC, equipment_key ASC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&items).Error
	return items, total, err
}

func (r *Repository) GetEquipment(ctx context.Context, key string) (*EquipmentDetail, error) {
	var item UnifiedEquipment
	if err := r.db.WithContext(ctx).First(&item, "equipment_key = ? OR prisma_id = ? OR startrack_id = ?", key, key, key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	maintenance := make([]UnifiedMaintenance, 0)
	if err := r.db.WithContext(ctx).Where("equipment_key = ?", item.EquipmentKey).Order("service_date DESC").Find(&maintenance).Error; err != nil {
		return nil, err
	}
	return &EquipmentDetail{Equipment: item, Maintenance: maintenance}, nil
}

func (r *Repository) ListAllMaintenance(ctx context.Context) ([]UnifiedMaintenance, error) {
	items := make([]UnifiedMaintenance, 0)
	err := r.db.WithContext(ctx).Order("service_date DESC").Find(&items).Error
	return items, err
}

func (r *Repository) ListTypes(ctx context.Context) ([]EquipmentType, error) {
	items := make([]EquipmentType, 0)
	err := r.db.WithContext(ctx).Where("active = ?", true).Order("name ASC").Find(&items).Error
	return items, err
}

func (r *Repository) GetRequest(ctx context.Context, id string) (*UnifiedRequest, error) {
	var item UnifiedRequest
	if err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *Repository) ListRequests(ctx context.Context, page, pageSize int, queryText, status, equipmentType, requester string) ([]UnifiedRequest, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	query := r.db.WithContext(ctx).Model(&UnifiedRequest{})
	if status != "" {
		query = query.Where("LOWER(status) = LOWER(?)", strings.TrimSpace(status))
	}
	if equipmentType != "" {
		query = query.Where("LOWER(type) = LOWER(?) OR equipment_type_code = ?", strings.TrimSpace(equipmentType), strings.ToUpper(strings.TrimSpace(equipmentType)))
	}
	if requester != "" {
		query = query.Where("requester ILIKE ?", "%"+strings.TrimSpace(requester)+"%")
	}
	if queryText != "" {
		pattern := "%" + strings.TrimSpace(queryText) + "%"
		query = query.Where("project ILIKE ? OR type ILIKE ? OR requester ILIKE ? OR machinery ILIKE ?", pattern, pattern, pattern, pattern)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := make([]UnifiedRequest, 0)
	err := query.Order("start_date DESC, id ASC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&items).Error
	return items, total, err
}

func (r *Repository) ListAssignments(ctx context.Context, page, pageSize int, status, search string) ([]UnifiedAssignment, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	query := r.db.WithContext(ctx).Model(&UnifiedAssignment{})
	if status != "" {
		query = query.Where("LOWER(status) = LOWER(?)", strings.TrimSpace(status))
	}
	if search != "" {
		pattern := "%" + strings.TrimSpace(search) + "%"
		query = query.Where("request_id ILIKE ? OR equipment_key ILIKE ? OR task_external_id ILIKE ? OR title ILIKE ? OR assignee ILIKE ?", pattern, pattern, pattern, pattern, pattern)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := make([]UnifiedAssignment, 0)
	err := query.Order("scheduled_date DESC, id ASC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&items).Error
	return items, total, err
}

func (r *Repository) GetAssignmentByRequest(ctx context.Context, requestID string) (*UnifiedAssignment, error) {
	var item UnifiedAssignment
	if err := r.db.WithContext(ctx).Where("request_id = ?", requestID).Order("synced_at DESC").First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *Repository) ListConflicts(ctx context.Context, page, pageSize int) ([]SyncConflict, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	var total int64
	if err := r.db.WithContext(ctx).Model(&SyncConflict{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := make([]SyncConflict, 0)
	err := r.db.WithContext(ctx).Order("detected_at DESC, entity_key ASC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&items).Error
	return items, total, err
}

func (r *Repository) GetSyncState(ctx context.Context) (*SyncState, error) {
	var state SyncState
	if err := r.db.WithContext(ctx).First(&state, "id = ?", 1).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &SyncState{ID: 1, Status: "NOT_SYNCED"}, nil
		}
		return nil, err
	}
	return &state, nil
}

func (r *Repository) Dashboard(ctx context.Context) (*Dashboard, error) {
	var result Dashboard
	if err := r.db.WithContext(ctx).Model(&UnifiedEquipment{}).Count(&result.EquipmentTotal).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&UnifiedEquipment{}).Where("available = ?", true).Count(&result.EquipmentAvailable).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&UnifiedRequest{}).Where("LOWER(status) = ?", "pendiente").Count(&result.RequestsPending).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&UnifiedAssignment{}).Where("status = ?", "ACTIVE").Count(&result.AssignmentsActive).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&SyncConflict{}).Count(&result.Conflicts).Error; err != nil {
		return nil, err
	}
	state, err := r.GetSyncState(ctx)
	if err != nil {
		return nil, err
	}
	result.LastSyncedAt = state.LastSucceededAt
	return &result, nil
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
