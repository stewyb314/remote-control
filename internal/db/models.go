package db

type Execution struct {
	ID     string `gorm:"primaryKey"`
	Status string
	Command string
	CreatedAt string
	Args []string `gorm:"type:json"`
}	
	