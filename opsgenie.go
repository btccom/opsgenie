// Package opsgenie provides basic heartbeat support.
//
// Example usage:
//    heartbeat := opsgenie.Heartbeat{
// 	      TeamName: "ops_team", // Optional. ops_team will be used by default.
// 	      Interval: 2, // Optional. 60 seconds by default.
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

const apiUrl = "https://api.opsgenie.com/v2"

var (
	apiKey  string
	appName string
)

func Configure(uApiKey, uAppName string) {
	apiKey = fmt.Sprintf("GenieKey %s", uApiKey)
	appName = uAppName
}

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

type Heartbeat struct {
	TeamName string        // Team name to assign. Should exists.
	Interval time.Duration // Interval in seconds to perform heartbeat requests

	quit chan int
}

type alert struct {
	Message     string `json:"message"`
	Description string `json:"description"`
	Entity      string `json:"entity"`
	Priority    string `json:"priority"`
}

func sendHeartbeat(h *Heartbeat) {
	url := fmt.Sprintf("%s/heartbeats/%s/ping", apiUrl, appName)

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
		Name:         appName,
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

	url := fmt.Sprintf("%s/heartbeats", apiUrl)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Add("Authorization", apiKey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending createHeartbeat request")
	} else {
		resp.Body.Close()
	}
}

func ReportError(uerr error, stack []byte) {
	alert := alert{
		Message:     uerr.Error(),
		Description: fmt.Sprintf("%s", stack),
		Entity:      appName,
		Priority:    "P1",
	}

	body, _ := json.Marshal(alert)

	url := fmt.Sprintf("%s/alerts", apiUrl)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Add("Authorization", apiKey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending alert")
	} else {
		resp.Body.Close()
	}
}

// Start sending heartbeat request.
func (h *Heartbeat) Start() error {
	if len(apiKey) == 0 {
		return errors.New("Please provide OpsGenie apikey")
	}
	if len(h.TeamName) == 0 {
		h.TeamName = "ops_team"
	}
	if h.Interval < 1 {
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
