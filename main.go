package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	ics "github.com/arran4/golang-ical"
)

type Match struct {
	Tournament string
	HomeTeam   string
	AwayTeam   string
	StartTime  time.Time
	Location   string
}

var events map[int64]Match
var data string

const Interval = time.Hour * 24
const Port = 8080

func main() {
	events = make(map[int64]Match)

	// Parse the endpoint and add the events to calendar every hour
	go func() {
		// Start a ticker that runs every 24 hours
		for {
			fmt.Printf("%s | Collecting events\n", time.Now().Format(time.RFC3339))
			evs, err := CollectEvents()
			if err != nil {
				fmt.Println("Error collecting events:", err)
				continue
			}
			fmt.Printf("%s | Collected %d events", time.Now().Format(time.RFC3339), len(evs))

			// Setup the calendar
			cal := ics.NewCalendar()
			cal.SetMethod(ics.MethodPublish)
			for _, ev := range evs {
				match := Match{
					Tournament: ev.Tournament.Name,
					HomeTeam:   ev.HomeTeam.Name,
					AwayTeam:   ev.AwayTeam.Name,
					StartTime:  time.Unix(int64(ev.StartTime), 0),
					Location:   fmt.Sprintf("%s, %s", ev.Venue.City.Name, ev.Venue.Stadium.Name),
				}
				events[ev.ID] = match

				event := cal.AddEvent(fmt.Sprintf("%s - %s", match.HomeTeam, match.AwayTeam))
				event.SetDtStampTime(time.Now())
				event.SetStartAt(match.StartTime)
				event.SetEndAt(match.StartTime.Add(time.Hour * 2))
				event.SetSummary(fmt.Sprintf("%s - %s", match.HomeTeam, match.AwayTeam))
				event.SetDescription(match.Tournament)
				event.SetLocation(match.Location)
			}
			data = cal.Serialize()
			time.Sleep(Interval)
		}
	}()

	// Start a web server with one endpoint
	// that returns the calendar
	go func() {
		http.HandleFunc("/calendar", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/calendar")
			w.Header().Set("Content-Disposition", "attachment; filename=calendar.ics")
			w.Write([]byte(data))
		})
		fmt.Printf("Starting to serve at port %d\n", Port)
		_ = http.ListenAndServe(":"+strconv.Itoa(Port), nil)
	}()
	select {}
}
