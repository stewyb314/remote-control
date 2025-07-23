package db

type DB interface{
	GetExecution(id string) (*Execution, error)
	CreateExecution(execution Execution) error
	Migrate() error
}