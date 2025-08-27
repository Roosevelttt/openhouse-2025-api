package models

import "time"

//import "gorm.io/gorm"

type User struct {
	NRP    string `gorm:"primaryKey;column:nrp" db:"nrp" json:"nrp"`
	Name   string `gorm:"column:name" db:"name" json:"name"`
	LineID string `gorm:"column:line_id" db:"line_id" json:"line_id"`
	Phone  string `gorm:"column:phone" db:"phone" json:"phone"`
}

type Ukm struct {
	ID          string `gorm:"primaryKey;column:id" db:"id" json:"id"`
	Name        string `gorm:"column:name" db:"name" json:"name"`
	Slug        string `gorm:"column:slug" db:"slug" json:"slug"`
	CurrentSlot int    `gorm:"column:current_slot" db:"current_slot" json:"current_slot"`
	MaxSlot     int    `gorm:"column:max_slot" db:"max_slot" json:"max_slot"`
	RegistFee   int    `gorm:"column:regist_fee" db:"regist_fee" json:"regist_fee"`
}

type Division struct {
	ID        string     `gorm:"primaryKey;column:id" db:"id" json:"id"`
	Name      string     `gorm:"column:name" db:"name" json:"name"`
	Slug      string     `gorm:"column:slug" db:"slug" json:"slug"`
	CreatedAt *time.Time `gorm:"column:created_at" db:"created_at" json:"created_at,omitempty"`
	UpdatedAt *time.Time `gorm:"column:updated_at" db:"updated_at" json:"updated_at,omitempty"`
}

type Admin struct {
	ID         string     `gorm:"primaryKey;column:id" db:"id" json:"id"`
	UkmID      *string    `gorm:"column:ukm_id" db:"ukm_id" json:"ukm_id,omitempty"`
	DivisionID *string    `gorm:"column:division_id" db:"division_id" json:"division_id,omitempty"`
	Name       string     `gorm:"column:name" db:"name" json:"name"`
	NRP        string     `gorm:"column:nrp" db:"nrp" json:"nrp"`
	Field      string     `gorm:"column:field" db:"field" json:"field"`
	CreatedAt  *time.Time `gorm:"column:created_at" db:"created_at" json:"created_at,omitempty"`
	UpdatedAt  *time.Time `gorm:"column:updated_at" db:"updated_at" json:"updated_at,omitempty"`
}

type DetailRegistration struct {
	ID               string     `gorm:"primaryKey;column:id" db:"id" json:"id"`
	NRP              string     `gorm:"column:nrp" db:"nrp" json:"nrp"`
	UkmID            string     `gorm:"column:ukm_id" db:"ukm_id" json:"ukm_id"`
	Payment          string     `gorm:"column:payment" db:"payment" json:"payment"`
	Code             string     `gorm:"column:code" db:"code" json:"code"`
	DriveURL         string     `gorm:"column:drive_url" db:"drive_url" json:"drive_url"`
	FileValidated    int        `gorm:"column:file_validated" db:"file_validated" json:"file_validated"`
	PaymentValidated int        `gorm:"column:payment_validated" db:"payment_validated" json:"payment_validated"`
	CreatedAt        *time.Time `gorm:"column:created_at" db:"created_at" json:"created_at,omitempty"`
}

type Participant struct {
	ID               string     `db:"id" json:"id"`
	NRP              string     `db:"nrp" json:"nrp"`
	Name             string     `db:"name" json:"name"`
	LineID           string     `db:"line_id" json:"line_id"`
	Phone            string     `db:"phone" json:"phone"`
	UkmId            string     `db:"ukm_id" json:"ukm_id"`
	UkmName          string     `db:"ukm_name" json:"ukm_name"`
	Payment          *string    `db:"payment" json:"payment"`
	FileValidated    int        `gorm:"column:file_validated" db:"file_validated" json:"file_validated"`
	PaymentValidated int        `db:"payment_validated" json:"payment_validated"`
	CreatedAt        *time.Time `db:"created_at" json:"created_at,omitempty"`
	IsInvited        int        `db:"isInvited" json:"is_invited"`
}
