package services

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sync"
	pb "github.com/stewyb314/remote-control/protos"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stewyb314/remote-control/internal/db"
	"gorm.io/datatypes"
)

type job struct {
	cancel func()
}

type Jobs struct {
	jobs map[string]job
	db db.DB
	log *logrus.Entry
	doneChan chan JobDone
}

type JobDone struct {
	id string
	status int32
	ExitCode int32
}



func NewJobs(db db.DB, log *logrus.Entry) *Jobs {
	j := &Jobs{
		jobs: make(map[string]job),
		db: db,
		log: log,
	}
	j.doneChan = make(chan JobDone)
	j.monitorJobs()
	return j
}	

func (j *Jobs) monitorJobs() {
	go func() {
		for done := range j.doneChan {
			j.log.Infof("Job %s finished with status %d and exit code %d", done.id, done.status, done.ExitCode)
			delete(j.jobs, done.id)
			exec, err := j.db.GetExecution(done.id)
			if err != nil {
				j.log.Errorf("Failed to get execution for job %s: %v", done.id, err)
				continue
			}
			exec.Status = done.status
			exec.ExitCode = done.ExitCode
			if err := j.db.UpdateExecution(*exec); err != nil {
				j.log.Errorf("Failed to update execution for job %s: %v", done.id, err)
				continue
			}
		}
		j.log.Infof("Done channel closed, stopping job monitoring")
	}()
	j.log.Infof("Done Monitoring jobs")
}

func (j *Jobs) NewJob(command string, args []string) (string, error){
	id := uuid.New().String()
	file := "jobs/" + id + ".txt"
	fw, err  := fileWrite(file)
	if err != nil {
		j.log.Errorf("Failed to create file %s: %v", file, err)
		return "", fmt.Errorf("failed to create file %s: %v", file, err)
	}
	a, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("failed to marshal args: %v", err)
	}
	ctx, cancel  := context.WithCancel(context.Background())
	
	newJob := job{
		cancel: cancel,
	}


	cmd := db.Execution{
		Command: command,
		Args:    datatypes.JSON(a),
		ID:     id,
		Status: int32(pb.State_RUNNING),
		Output: file,
	}
	j.log.Infof("Creating new job %s with command %+v", id, cmd)

	if err := j.db.CreateExecution(cmd); err != nil {
		j.log.Errorf("Failed to create execution: %v", err)
		return "", fmt.Errorf("failed to create execution: %v", err)
	}
	j.jobs[id] = newJob
	j.startJob(ctx, command, args, fw, id)
	return id, nil
}

func (j *Jobs) StopJob(id string) error {
	exec, err := j.db.GetExecution(id)
	if err != nil {
		return fmt.Errorf("failed to get execution for job ID %s: %v", id, err)
	}
	if exec.Status == int32(pb.State_COMPLETE) || exec.Status == int32(pb.State_STOPPED) {
		return nil
	}
	job, ok := j.jobs[id]
	if !ok {
		return fmt.Errorf("no job found with ID %s", id)
	}
	job.cancel()
	delete(j.jobs, id)
	j.log.Infof("Job %s stopped", id)
	return nil
}

func (j *Jobs) startJob(ctx context.Context, cmd string, args[]string, output *bufio.Writer, id string) {
	
	j.log.Infof("Starting job %s with args %v", cmd, args)
	finished := make(chan struct{})

	execCmd := exec.CommandContext(ctx, cmd, args...)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {	
		defer output.Flush()

		execCmd.Stdout = output
		execCmd.Stderr =  output
		err := execCmd.Start()	
		wg.Done()
		if err != nil {
			j.doneChan <- JobDone{status: int32(pb.State_ERROR), ExitCode: -1, id: id}
			return
		}
			select {
			case <-ctx.Done():
				if err := execCmd.Process.Kill(); err != nil {
					j.doneChan <- JobDone{status: int32(pb.State_ERROR), ExitCode: -1, id: id}
				} else {
					j.doneChan <- JobDone{status: int32(pb.State_STOPPED), ExitCode: 0, id: id}
				}
			case <-finished:
				j.doneChan <- JobDone{status: int32(pb.State_COMPLETE), ExitCode: int32(execCmd.ProcessState.ExitCode()), id: id }
			}
	}()
	// ensure that the execCmd.Start() has been called before we wait for it
	wg.Wait()
	go func ()  {
		defer close(finished)
		execCmd.Wait()		
	}()

}

func fileWrite(file string) (*bufio.Writer, error){
	f, err := os.Create(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %s: %v", file, err)
	}
	return bufio.NewWriter(f), nil
}