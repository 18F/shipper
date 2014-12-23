package main

import (
	"log"
	"time"
)

func Run(config *Config) {
	log.Println("Running shipper with an interval of ", config.Interval, " seconds")
	c := time.Tick(time.Duration(config.Interval) * time.Second)
	for _ = range c {
		dep, err := config.Backend.FindNewDeployment()
		if err != nil {
			// There was an error lets print it to the log and continue
			log.Println(err)
			continue
		}

		if dep != nil {
			if err := config.Backend.UpdateStatus(dep, "pending"); err != nil {
				// Couldn't update status lets try next tick
				log.Println(err)
				continue
			}

			// Run deployment
			err := Deploy(config, dep)

			if err != nil {
				// There was an error deploying
				log.Println("Error deploying")
				if err := config.Backend.UpdateStatus(dep, "error"); err != nil {
					// Couldn't update status lets try next tick
					log.Println(err)
				}
				continue
			}

			// Successful deploy
			log.Println("Set to success")
			if err := config.Backend.UpdateStatus(dep, "success"); err != nil {
				// Couldn't update status lets try next tick
				log.Println(err)
				continue
			}

		}
	}
}
