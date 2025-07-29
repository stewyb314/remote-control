package main

import (
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stewyb314/remote-control/internal/agent"
	"github.com/stewyb314/remote-control/internal/config"
	"github.com/stewyb314/remote-control/internal/db"
	"github.com/stewyb314/remote-control/internal/services"
)

func main() {
	log := logrus.New().WithField("request_id", uuid.New().String()) 
	log.Logger.Formatter = &logrus.JSONFormatter{}
	conf := config.NewAgentConfig()
	mysql, err  := db.NewMySQL(conf.DbConfig)
	if err != nil {
		log.Infof("Failed to connect to MySQL: %v", err)
	}

	if err := mysql.Migrate(); err != nil {
		log.Infof("Failed to migrate database: %v", err)
	}
	jobs := services.NewJobs(mysql, log)
	a := agent.New(log, "0.0.0.0", 50051, nil, mysql, jobs)
	log.Infof("Starting agent")
	err = a.StartAgent()
	if err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}
}