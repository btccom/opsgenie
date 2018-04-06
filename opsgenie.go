// Package ospgenie provides basic heartbeat support.
//
// Example usage:
//    heartbeat := opsgenie.Heartbeat{
// 	      ApiKey:   "your-api-key-with-configuration-access",
// 	      PingName: "service-name",
// 	      TeamName: "ops_team",
// 	      Interval: 2,
//    }
//    err = heartbeat.Start()
//    if err != nil {
// 	      return err
//    }
//    defer heartbeat.Stop()

package opsgenie

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const apiUrl = "https://api.opsgenie.com/v2/heartbeats/"

type newPing struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	IntervalUnit string `json:"intervalUnit"`
	Interval     int    `json:"interval"`
	Enabled      bool   `json:"enabled"`

	OwnerTeam struct {
		Name string `json:"name"`
	} `json:"ownerTeam"`
}

func sendHeartbeat(h *Heartbeat) {
	apiKey := fmt.Sprintf("GenieKey %s", h.ApiKey)
	url := fmt.Sprintf("%s/%s/ping", apiUrl, h.PingName)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending heartbeat request")
	} else {
		resp.Body.Close()
	}
}

func createHeartbeart(h *Heartbeat) {
	newPing := newPing{
		Name:         h.PingName,
		Description:  "",
		IntervalUnit: "minutes",
		Interval:     5,
		Enabled:      true,
		OwnerTeam: struct {
			Name string `json:"name"`
		}{
			Name: h.TeamName,
		},
	}
	body, _ := json.Marshal(newPing)

	apiKey := fmt.Sprintf("GenieKey %s", h.ApiKey)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", apiUrl, bytes.NewBuffer(body))
	req.Header.Add("Authorization", apiKey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending createHeartbeat request")
	} else {
		resp.Body.Close()
	}
}

type Heartbeat struct {
	ApiKey   string
	PingName string
	TeamName string        // Team name to assign. Should exists.
	Interval time.Duration // Interval in seconds to perform heartbeat requests

	quit chan int
}

func (h *Heartbeat) Start() error {
	if len(h.ApiKey) == 0 {
		return errors.New("Please provide OpsGenie apikey")
	}
	if len(h.TeamName) == 0 {
		h.TeamName = "ops_team"
	}
	if h.Interval <= 1 {
		h.Interval = 60
	}
	createHeartbeart(h)
	go sendHeartbeat(h)

	h.quit = make(chan int)

	go func(h *Heartbeat) {
		ticker := time.NewTicker(time.Second * h.Interval)
		for {
			select {
			case <-ticker.C:
				sendHeartbeat(h)
			case <-h.quit:
				return
			}
		}
	}(h)

	return nil
}

func (h *Heartbeat) Stop() {
	h.quit <- 0
}
