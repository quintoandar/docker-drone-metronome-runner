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
	status := resp.(metronome.JobStatus).Status

	log.WithFields(log.Fields{
		"job":    p.Job,
		"id":     id,
		"status": status,
	}).Info("waiting for job to finish")

	timeout := time.After(p.Timeout)
	tick := time.Tick(10 * time.Second)

	for {
		select {

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

			if err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("failed to get job status")
				return err
			}

			status := jobStatus.Status

			if status == "Completed" {
				log.WithFields(log.Fields{
					"job":    p.Job,
					"id":     id,
					"status": status,
				}).Info("job has completed successfuly")
				return nil
			}

			if status == "Failed" {
				log.WithFields(log.Fields{
					"job":    p.Job,
					"id":     id,
					"status": status,
				}).Error("job has failed")
				return err
			}

			log.WithFields(log.Fields{
				"job":    p.Job,
				"id":     id,
				"status": status,
			}).Debug()
		}
	}
}
