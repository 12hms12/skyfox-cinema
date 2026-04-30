package model

type Slot struct {
	Id        int    `json:"id" gorm:"primaryKey"`
	Name      string `json:"name" gorm:"uniqueIndex:idx_slot_name"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

type Tabler interface {
	TableName() string
}

func (Slot) TableName() string {
	return "slot"
}
