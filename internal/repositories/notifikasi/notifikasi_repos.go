package notifikasi

import (
	"gorm.io/gorm"

	"github.com/projsonal/gowms/internal/model"
	"github.com/projsonal/gowms/pkg/utils"
)

type repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(n *model.Notification) error {
	return r.db.Create(n).Error
}

func scopeVisible(db *gorm.DB, userID uint, userRole string) *gorm.DB {
	return db.
		Where("user_id = ? OR role_target = ? OR role_target = ?", userID, userRole, "all").
		Where("id NOT IN (SELECT notification_id FROM notification_dismissed WHERE user_id = ?)", userID)
}

func (r *repository) List(userID uint, userRole string, p utils.PaginationParams) ([]Row, int64, error) {
	base := scopeVisible(r.db.Model(&model.Notification{}), userID, userRole)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []model.Notification
	q := scopeVisible(r.db.Model(&model.Notification{}), userID, userRole).Order("created_at DESC")
	q = p.Apply(q)
	if err := q.Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	if len(rows) == 0 {
		return []Row{}, total, nil
	}

	ids := make([]uint, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}
	var readIDs []uint
	if err := r.db.Model(&model.NotificationRead{}).
		Where("user_id = ? AND notification_id IN ?", userID, ids).
		Pluck("notification_id", &readIDs).Error; err != nil {
		return nil, 0, err
	}
	readSet := make(map[uint]bool, len(readIDs))
	for _, id := range readIDs {
		readSet[id] = true
	}

	out := make([]Row, len(rows))
	for i, row := range rows {
		out[i] = Row{Notification: row, IsRead: readSet[row.ID]}
	}
	return out, total, nil
}

func (r *repository) UnreadCount(userID uint, userRole string) (int64, error) {
	var count int64
	err := scopeVisible(r.db.Model(&model.Notification{}), userID, userRole).
		Where("id NOT IN (SELECT notification_id FROM notification_reads WHERE user_id = ?)", userID).
		Count(&count).Error
	return count, err
}

func (r *repository) MarkRead(notificationID, userID uint) error {
	return r.db.Exec(
		`INSERT INTO notification_reads (notification_id, user_id, read_at)
		 VALUES (?, ?, NOW())
		 ON CONFLICT (notification_id, user_id) DO NOTHING`,
		notificationID, userID,
	).Error
}

func (r *repository) MarkAllRead(userID uint, userRole string) error {
	var ids []uint
	if err := scopeVisible(r.db.Model(&model.Notification{}), userID, userRole).
		Pluck("id", &ids).Error; err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	return r.db.Exec(
		`INSERT INTO notification_reads (notification_id, user_id, read_at)
		 SELECT unnest(?::int[]), ?, NOW()
		 ON CONFLICT (notification_id, user_id) DO NOTHING`,
		ids, userID,
	).Error
}

func (r *repository) Dismiss(notificationID, userID uint) error {
	return r.db.Exec(
		`INSERT INTO notification_dismissed (notification_id, user_id, dismissed_at)
		 VALUES (?, ?, NOW())
		 ON CONFLICT (notification_id, user_id) DO NOTHING`,
		notificationID, userID,
	).Error
}
