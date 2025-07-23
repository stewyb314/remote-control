package main

import (

	"github.com/sirupsen/logrus"
	"github.com/stewyb314/remote-control/internal/agent"
	"github.com/stewyb314/remote-control/internal/config"
	"github.com/stewyb314/remote-control/internal/db"
)

func main() {
	log := logrus.New()
	conf := config.NewAgentConfig()
	mysql, err  := db.NewMySQL(conf.DbConfig)
	if err != nil {
		log.Infof("Failed to connect to MySQL: %v", err)
	} else {
	if err := mysql.Migrate(); err != nil {
		log.Infof("Failed to migrate database: %v", err)
	}

	} 

	a := agent.New(log, "0.0.0.0", 50051, nil, mysql)
	err = a.StartAgent()
	if err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}
}