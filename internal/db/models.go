package db

import (

	"gorm.io/datatypes"
)
type Execution struct {
	ID     string `gorm:"primaryKey"`
	Status int32 
	Command string
	CreatedAt int64 `gorm:"autoCreateTime"` 
	UpdatedAt int64 `gorm:"autoUpdateTime"`
	Output string `gorm:"type:longtext"`
	ExitCode int32
	Args datatypes.JSON `gorm:"type:json"`
}	
	