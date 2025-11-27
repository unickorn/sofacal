package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"

	"golang.org/x/net/proxy"
)

const useProxy = true
const proxyAddr = "127.0.0.1:1080"

const eventsEndpoint = "https://www.sofascore.com/api/v1/team/%d/events/next/0"
const eventEndpoint = "https://www.sofascore.com/api/v1/event/%d"

var teams []int
var httpClient *http.Client

func init() {
	teams = []int{
		3052, // Fenerbahce Football
		3514, // Fenerbahce Basketball
		4700, // Turkiye Football
		6253, // Turkiye Basketball
	}

	if !useProxy {
		httpClient = &http.Client{}
		return
	}
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		fmt.Fprintln(os.Stderr, "can't connect to the proxy:", err)
		os.Exit(1)
	}
	httpTransport := &http.Transport{}
	httpClient = &http.Client{Transport: httpTransport}
	httpTransport.DialContext = func(_ context.Context, network, addr string) (net.Conn, error) {
		return dialer.Dial(network, addr)
	}
}

type Event struct {
	ID         int64  `json:"id"`
	Slug       string `json:"slug"`
	Tournament struct {
		Name string `json:"name"`
	} `json:"tournament"`
	HomeTeam struct {
		Name string `json:"name"`
	} `json:"homeTeam"`
	AwayTeam struct {
		Name string `json:"name"`
	} `json:"awayTeam"`
	StartTime uint64 `json:"startTimestamp"`
	Venue     struct {
		City struct {
			Name string `json:"name"`
		} `json:"city"`
		Stadium struct {
			Name string `json:"name"`
		} `json:"stadium"`
	} `json:"venue"`
}

func getEventIDs() ([]int64, error) {
	var ids []int64
	for _, team := range teams {
		rsp, err := httpClient.Get(fmt.Sprintf(eventsEndpoint, team))
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()

		// Parse the response into a struct.
		var eventIDs struct {
			Events []struct {
				ID int64 `json:"id"`
			} `json:"events"`
		}
		err = json.NewDecoder(rsp.Body).Decode(&eventIDs)
		if err != nil {
			return nil, err
		}
		// Return the IDs.
		for _, e := range eventIDs.Events {
			ids = append(ids, e.ID)
		}
	}
	return ids, nil
}

func getEvent(id int64) (Event, error) {
	rsp, err := httpClient.Get(fmt.Sprintf(eventEndpoint, id))
	if err != nil {
		return Event{}, err
	}
	defer rsp.Body.Close()

	// Parse the response into a struct.
	var event struct {
		Event Event `json:"event"`
	}
	err = json.NewDecoder(rsp.Body).Decode(&event)
	if err != nil {
		return Event{}, err
	}
	return event.Event, nil
}

// CollectEvents collects all events from the API.
func CollectEvents() ([]Event, error) {
	ids, err := getEventIDs()
	if err != nil {
		return nil, err
	}
	var evs []Event
	for _, id := range ids {
		event, err := getEvent(id)
		if err != nil {
			return nil, err
		}
		evs = append(evs, event)
	}
	return evs, nil
}
