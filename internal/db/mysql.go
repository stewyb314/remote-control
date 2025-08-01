package db

import (
	"fmt"

	"github.com/stewyb314/remote-control/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type MySQL struct {
	db *gorm.DB
}

func NewMySQL(conf config.DbConfig) (*MySQL, error) {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", conf.User, conf.Password, conf.Host, conf.Port, conf.Database)
	databaseConnection, err := gorm.Open(mysql.Open(connectionString), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err!=nil{
		return nil, err
	}

	return &MySQL{
		db: databaseConnection,
	}, nil

}

func (m MySQL) GetExecution(id string) (*Execution, error) {
	var execution Execution
	tx := m.db.First(&execution, "id = ?", id)
	return &execution, tx.Error
}
func (m MySQL) CreateExecution(execution Execution) error {
	tx := m.db.Create(&execution)
	if tx.Error != nil {
		return fmt.Errorf("error creating execution: %v", tx.Error)
	}

	return nil
}	

func (m MySQL) UpdateExecution(exec Execution) error {
	tx := m.db.Save(&exec)
	return tx.Error
}

func (m MySQL) Migrate() error {
	err := m.db.AutoMigrate(
		&Execution{},
	)
	if err != nil {
		return fmt.Errorf("error migrating database: %v", err)
	}

	return nil
}	