package models

import "time"

type User struct {
	NRP    string `db:"nrp" json:"nrp"`
	Name   string `db:"name" json:"name"`
	LineID string `db:"line_id" json:"line_id"`
	Phone  string `db:"phone" json:"phone"`
}

type Ukm struct {
	ID          string `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	Slug        string `db:"slug" json:"slug"`
	CurrentSlot int    `db:"current_slot" json:"current_slot"`
	MaxSlot     int    `db:"max_slot" json:"max_slot"`
	RegistFee   int    `db:"regist_fee" json:"regist_fee"`
}

type Division struct {
	ID        string     `db:"id" json:"id"`
	Name      string     `db:"name" json:"name"`
	Slug      string     `db:"slug" json:"slug"`
	CreatedAt *time.Time `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt *time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

type Admin struct {
	ID         string     `db:"id" json:"id"`
	UkmID      *string    `db:"ukm_id" json:"ukm_id,omitempty"`
	DivisionID *string    `db:"division_id" json:"division_id,omitempty"`
	Name       string     `db:"name" json:"name"`
	NRP        string     `db:"nrp" json:"nrp"`
	Field      string     `db:"field" json:"field"`
	CreatedAt  *time.Time `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt  *time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

type DetailRegistration struct {
	ID               string     `db:"id" json:"id"`
	NRP              string     `db:"nrp" json:"nrp"`
	UkmID            string     `db:"ukm_id" json:"ukm_id"`
	Payment          int        `db:"payment" json:"payment"`
	Code             string     `db:"code" json:"code"`
	DriveURL         string     `db:"drive_url" json:"drive_url"`
	FileValidated    int        `db:"file_validated" json:"file_validated"`
	PaymentValidated int        `db:"payment_validated" json:"payment_validated"`
	CreatedAt        *time.Time `db:"created_at" json:"created_at,omitempty"`
}

type Participant struct {
	ID               string     `db:"id" json:"id"`
	NRP              string     `db:"nrp" json:"nrp"`
	Name             string     `db:"name" json:"name"`
	LineID           string     `db:"line_id" json:"line_id"`
	Phone            string     `db:"phone" json:"phone"`
	UkmName          string     `db:"ukm_name" json:"ukm_name"`
	Payment          *string    `db:"payment" json:"payment"`
	FileValidated    int        `db:"file_validated" json:"file_validated"`
	PaymentValidated int        `db:"payment_validated" json:"payment_validated"`
	CreatedAt        *time.Time `db:"created_at" json:"created_at,omitempty"`
	IsInvited        int        `db:"isInvited" json:"is_invited"`
}
