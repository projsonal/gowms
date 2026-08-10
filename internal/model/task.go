package model

import "time"

// Task — modul Task Management (penugasan internal admin/super_admin ke
// user tertentu, biasanya karyawan). Status "terlambat" TIDAK disimpan di
// kolom terpisah — dihitung on-the-fly (DueDate sudah lewat & Status masih
// "baru"/"proses") supaya tidak ada dua sumber kebenaran yang bisa saling
// tidak sinkron; lihat repositories/task IsOverdue().
type Task struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Title       string     `json:"title" gorm:"size:150;not null"`
	Description string     `json:"description" gorm:"size:500"`
	AssignedTo  uint       `json:"assigned_to" gorm:"not null;index"`
	Assignee    *User      `json:"assignee,omitempty" gorm:"foreignKey:AssignedTo"`
	AssignedBy  uint       `json:"assigned_by" gorm:"not null"`
	Assigner    *User      `json:"assigner,omitempty" gorm:"foreignKey:AssignedBy"`
	DueDate     time.Time  `json:"due_date" gorm:"not null"`
	Priority    string     `json:"priority" gorm:"size:10;not null;default:'sedang'"` // rendah | sedang | tinggi
	Status      string     `json:"status" gorm:"size:10;not null;default:'baru';index"` // baru | proses | selesai
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Task) TableName() string { return "tasks" }

func (t *Task) IsOverdue(now time.Time) bool {
	return t.Status != "selesai" && now.After(t.DueDate)
}
