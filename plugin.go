package main

import (
	"errors"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/adobe-platform/go-metronome/metronome"
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

	client, err := metronome.NewClient(metronome.Config{
		URL:            p.URL,
		AuthToken:      p.Token,
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
	runID := resp.(string)

	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("failed to start job")
	}

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
			jobStatus, err := client.StatusJob(p.Job, runID)

			if err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("failed to get job status")
				return err
			}

			status := jobStatus.Status

			if status == "Completed" {
				log.WithFields(log.Fields{
					"url":    p.URL,
					"job":    p.Job,
					"status": status,
				}).Info("job has completed sucesfully")
				return nil
			}

			if status == "Failed" {
				log.WithFields(log.Fields{
					"url":    p.URL,
					"job":    p.Job,
					"status": status,
				}).Error("job has failed")
				return err
			}

			log.WithFields(log.Fields{
				"url":    p.URL,
				"job":    p.Job,
				"status": status,
			}).Debug()
		}
	}
}
