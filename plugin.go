package main

import (
	"errors"
	"net/url"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/fbcbarbosa/go-metronome/metronome"
)

// Plugin defines the beanstalk plugin parameters.
type Plugin struct {
	URL     string
	Token   string
	Job     string
	Timeout time.Duration
}

// Exec runs the plugin
func (p *Plugin) Exec() error {

	log.WithFields(log.Fields{
		"url": p.URL,
		"job": p.Job,
	}).Info("attempting to start job")

	u, err := url.Parse(p.URL)

	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to paser dcos url")
		return err
	}

	u.Path = path.Join(u.Path, "service/metronome")

	client, err := metronome.NewClient(metronome.Config{
		URL:            u.String(),
		AuthToken:      "token=" + p.Token,
		Debug:          false,
		RequestTimeout: 5,
	})

	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to start Metronome client")
		return err
	}

	jobs, err := client.Jobs()

	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to retrieve available jobs")
	}

	for _, j := range *jobs {
		log.Debugf("found job %s", j.GetID())
	}

	resp, err := client.StartJob(p.Job)

	if err != nil {
		log.WithFields(log.Fields{
			"job": p.Job,
		}).Error("failed to start job")
		return err
	}

	id := resp.(metronome.JobStatus).ID

	log.WithFields(log.Fields{
		"job": p.Job,
		"id":  id,
	}).Info("job is starting")

	timeout := time.After(p.Timeout)
	tick := time.Tick(10 * time.Second)
	printTick := time.Tick(1 * time.Minute)

	for {
		select {

		case <-printTick:
			log.WithFields(log.Fields{
				"job": p.Job,
				"id":  id,
			}).Info("waiting for job to finish")

		case <-timeout:
			err := errors.New("timed out")

			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("failed to get job status")
				return err
			}

		case <-tick:
			jobStatus, err := client.StatusJob(p.Job, id)

			// if err, it means that the job is not running and might have finished
			if err != nil {

				log.WithFields(log.Fields{
					"err": err,
				}).Info("job is not running")

				job, err := client.GetJob(p.Job)

				if err != nil {
					log.WithFields(log.Fields{
						"error": err,
					}).Error("failed to get job status")
					return err
				}

				if job.History == nil {
					return errors.New("failed to get job history")
				}

				if job.History.SuccessfulFinishedRuns != nil {
					for _, run := range job.History.SuccessfulFinishedRuns {
						if run.ID == id {
							log.Info("job has completed successfully")
							return nil
						}
					}
				} else {
					log.Warn("no sucessful runs were found for this job")
				}

				if job.History.FailedFinishedRuns != nil {
					for _, run := range job.History.FailedFinishedRuns {
						if run.ID == id {
							return errors.New("job has failed")
						}
					}
				}

				return errors.New("run was not found on job history")
			}

			log.WithFields(log.Fields{
				"job":    p.Job,
				"id":     id,
				"status": jobStatus.Status,
			}).Debug()
		}
	}
}
