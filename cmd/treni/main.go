package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/emiliopalmerini/treni/internal/api/viaggiatreno"
	"github.com/emiliopalmerini/treni/internal/storage"
	"github.com/emiliopalmerini/treni/internal/storage/sqlc"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "train":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "error: train number required")
			os.Exit(1)
		}
		trainCmd(args[0])
	case "station":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "error: station code or name required")
			os.Exit(1)
		}
		stationCmd(args[0])
	case "search":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "error: search query required")
			os.Exit(1)
		}
		searchCmd(args[0])
	case "history":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "error: train number required")
			os.Exit(1)
		}
		historyCmd(args[0])
	case "record":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "error: train number required")
			os.Exit(1)
		}
		recordCmd(args[0])
	case "stats":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "error: train number required")
			os.Exit(1)
		}
		statsCmd(args[0])
	case "top":
		topCmd(args)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "error: unknown command %q\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`treni - Train tracking CLI

Usage:
  treni <command> [arguments]

Commands:
  train <number>     Get real-time status for a train
  station <code>     Get arrivals/departures for a station
  search <query>     Search for stations by name
  record <number>    Record current train delay to database
  history <number>   Get historical delays for a train
  stats <number>     Get statistics for a train
  top [delayed|reliable]  Show top delayed or reliable trains
  help               Show this help message

Examples:
  treni train 9311
  treni station S01700
  treni search Milano
  treni record 9311
  treni history 9311
  treni stats 9311
  treni top delayed
  treni top reliable`)
}

func getDB() (*storage.DB, *sqlc.Queries, error) {
	// Check for Treni Turso env vars first
	if os.Getenv("TRENI_DATABASE_URL") != "" {
		db, err := storage.New()
		if err != nil {
			return nil, nil, err
		}
		return db, sqlc.New(db.DB), nil
	}

	// Fall back to local SQLite
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, err
	}

	dbPath := filepath.Join(home, ".local", "share", "treni", "treni.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, nil, err
	}

	db, err := storage.NewLocal(dbPath)
	if err != nil {
		return nil, nil, err
	}

	// Run migrations
	if err := db.Migrate(); err != nil {
		db.Close()
		return nil, nil, err
	}

	return db, sqlc.New(db.DB), nil
}

func trainCmd(trainNumber string) {
	client := viaggiatreno.New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	train, err := client.GetTrain(ctx, trainNumber)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s %s\n", train.Category, train.Number)
	fmt.Printf("%s → %s\n", train.Origin, train.Destination)
	fmt.Printf("Status: %s\n", train.Status)
	if train.Delay > 0 {
		fmt.Printf("Delay: +%d min\n", train.Delay)
	} else if train.Delay < 0 {
		fmt.Printf("Delay: %d min (early)\n", train.Delay)
	} else {
		fmt.Println("Delay: On time")
	}
	if !train.LastUpdate.IsZero() {
		fmt.Printf("Last update: %s\n", train.LastUpdate.Format("15:04"))
	}

	if len(train.Stops) > 0 {
		fmt.Println("\nStops:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "Station\tArr\tDep\tDelay\tPlatform")
		fmt.Fprintln(w, "-------\t---\t---\t-----\t--------")
		for _, stop := range train.Stops {
			arr := "-"
			dep := "-"
			delay := "-"
			if !stop.ScheduledArrival.IsZero() {
				arr = stop.ScheduledArrival.Format("15:04")
			}
			if !stop.ScheduledDepart.IsZero() {
				dep = stop.ScheduledDepart.Format("15:04")
			}
			if stop.DepartureDelay != 0 {
				delay = fmt.Sprintf("%+d", stop.DepartureDelay)
			} else if stop.ArrivalDelay != 0 {
				delay = fmt.Sprintf("%+d", stop.ArrivalDelay)
			}
			platform := stop.Platform
			if platform == "" {
				platform = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", stop.StationName, arr, dep, delay, platform)
		}
		w.Flush()
	}
}

func stationCmd(stationCode string) {
	client := viaggiatreno.New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// If it doesn't look like a station code, search first
	if len(stationCode) < 3 || stationCode[0] != 'S' {
		stations, err := client.SearchStation(ctx, stationCode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if len(stations) == 0 {
			fmt.Fprintf(os.Stderr, "error: no stations found for %q\n", stationCode)
			os.Exit(1)
		}
		if len(stations) > 1 {
			fmt.Println("Multiple stations found:")
			for _, s := range stations {
				fmt.Printf("  %s - %s\n", s.Code, s.Name)
			}
			fmt.Println("\nUse the station code to get details.")
			return
		}
		stationCode = stations[0].Code
		fmt.Printf("Using station: %s (%s)\n\n", stations[0].Name, stations[0].Code)
	}

	station, err := client.GetStation(ctx, stationCode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Departures
	fmt.Println("DEPARTURES")
	if len(station.Departures) == 0 {
		fmt.Println("  No departures found")
	} else {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "Time\tTrain\tDestination\tDelay\tPlatform")
		fmt.Fprintln(w, "----\t-----\t-----------\t-----\t--------")
		for _, d := range station.Departures {
			delay := ""
			if d.Delay > 0 {
				delay = fmt.Sprintf("+%d", d.Delay)
			}
			platform := d.Platform
			if platform == "" {
				platform = "-"
			}
			fmt.Fprintf(w, "%s\t%s %s\t%s\t%s\t%s\n",
				d.ScheduledTime.Format("15:04"),
				d.TrainCategory,
				d.TrainNumber,
				d.Destination,
				delay,
				platform,
			)
		}
		w.Flush()
	}

	fmt.Println()

	// Arrivals
	fmt.Println("ARRIVALS")
	if len(station.Arrivals) == 0 {
		fmt.Println("  No arrivals found")
	} else {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "Time\tTrain\tOrigin\tDelay\tPlatform")
		fmt.Fprintln(w, "----\t-----\t------\t-----\t--------")
		for _, a := range station.Arrivals {
			delay := ""
			if a.Delay > 0 {
				delay = fmt.Sprintf("+%d", a.Delay)
			}
			platform := a.Platform
			if platform == "" {
				platform = "-"
			}
			fmt.Fprintf(w, "%s\t%s %s\t%s\t%s\t%s\n",
				a.ScheduledTime.Format("15:04"),
				a.TrainCategory,
				a.TrainNumber,
				a.Origin,
				delay,
				platform,
			)
		}
		w.Flush()
	}
}

func searchCmd(query string) {
	client := viaggiatreno.New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stations, err := client.SearchStation(ctx, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if len(stations) == 0 {
		fmt.Printf("No stations found for %q\n", query)
		return
	}

	fmt.Printf("Found %d stations:\n", len(stations))
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Code\tName")
	fmt.Fprintln(w, "----\t----")
	for _, s := range stations {
		fmt.Fprintf(w, "%s\t%s\n", s.Code, s.Name)
	}
	w.Flush()
}

func recordCmd(trainNumber string) {
	client := viaggiatreno.New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	train, err := client.GetTrain(ctx, trainNumber)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	db, queries, err := getDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	cancelled := train.Status == "cancelled"
	today := time.Now().Truncate(24 * time.Hour)

	err = queries.InsertDelayRecord(ctx, sqlc.InsertDelayRecordParams{
		TrainNumber:   train.Number,
		TrainCategory: sql.NullString{String: train.Category, Valid: train.Category != ""},
		Origin:        train.Origin,
		Destination:   train.Destination,
		Date:          today,
		Delay:         int64(train.Delay),
		Cancelled:     sql.NullBool{Bool: cancelled, Valid: true},
		Source:        sql.NullString{String: "viaggiatreno", Valid: true},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error recording: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Recorded: %s %s (%s → %s) delay: %+d min\n",
		train.Category, train.Number, train.Origin, train.Destination, train.Delay)
}

func historyCmd(trainNumber string) {
	db, queries, err := getDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	records, err := queries.GetDelayRecordsByTrain(ctx, trainNumber)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error querying: %v\n", err)
		os.Exit(1)
	}

	if len(records) == 0 {
		fmt.Printf("No history found for train %s\n", trainNumber)
		fmt.Println("Use 'treni record <number>' to start recording delays.")
		return
	}

	fmt.Printf("History for train %s (%d records):\n\n", trainNumber, len(records))
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Date\tRoute\tDelay\tStatus")
	fmt.Fprintln(w, "----\t-----\t-----\t------")
	for _, r := range records {
		status := "OK"
		if r.Cancelled.Valid && r.Cancelled.Bool {
			status = "CANCELLED"
		} else if r.Delay > 5 {
			status = "DELAYED"
		}
		delay := fmt.Sprintf("%+d min", r.Delay)
		fmt.Fprintf(w, "%s\t%s → %s\t%s\t%s\n",
			r.Date.Format("2006-01-02"), r.Origin, r.Destination, delay, status)
	}
	w.Flush()
}

func statsCmd(trainNumber string) {
	db, queries, err := getDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stats, err := queries.GetTrainStats(ctx, trainNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("No stats found for train %s\n", trainNumber)
			return
		}
		fmt.Fprintf(os.Stderr, "error querying: %v\n", err)
		os.Exit(1)
	}

	if stats.TotalTrips == 0 {
		fmt.Printf("No stats found for train %s\n", trainNumber)
		return
	}

	onTimeTrips := int64(0)
	if stats.OnTimeTrips.Valid {
		onTimeTrips = int64(stats.OnTimeTrips.Float64)
	}
	delayedTrips := int64(0)
	if stats.DelayedTrips.Valid {
		delayedTrips = int64(stats.DelayedTrips.Float64)
	}
	cancelledTrips := int64(0)
	if stats.CancelledTrips.Valid {
		cancelledTrips = int64(stats.CancelledTrips.Float64)
	}

	onTimeRate := float64(onTimeTrips) / float64(stats.TotalTrips) * 100

	fmt.Printf("Statistics for train %s:\n\n", trainNumber)
	fmt.Printf("Total trips:     %d\n", stats.TotalTrips)
	fmt.Printf("On time:         %d (%.1f%%)\n", onTimeTrips, onTimeRate)
	fmt.Printf("Delayed:         %d\n", delayedTrips)
	fmt.Printf("Cancelled:       %d\n", cancelledTrips)
	if stats.AverageDelay.Valid {
		fmt.Printf("Average delay:   %.1f min\n", stats.AverageDelay.Float64)
	}
	if stats.MaxDelay != nil {
		if v, ok := stats.MaxDelay.(int64); ok {
			fmt.Printf("Max delay:       %d min\n", v)
		} else if v, ok := stats.MaxDelay.(float64); ok {
			fmt.Printf("Max delay:       %.0f min\n", v)
		}
	}
}

func topCmd(args []string) {
	subCmd := "delayed"
	if len(args) > 0 {
		subCmd = args[0]
	}

	db, queries, err := getDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	to := time.Now().Truncate(24 * time.Hour)
	from := to.AddDate(0, 0, -30)

	switch subCmd {
	case "delayed":
		trains, err := queries.GetMostDelayedTrains(ctx, sqlc.GetMostDelayedTrainsParams{
			FromDate:   from,
			ToDate:     to,
			LimitCount: 10,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error querying: %v\n", err)
			os.Exit(1)
		}

		if len(trains) == 0 {
			fmt.Println("No data found")
			return
		}

		fmt.Println("Most delayed trains (last 30 days):")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "Train\tRoute\tTrips\tAvg Delay\tMax Delay")
		fmt.Fprintln(w, "-----\t-----\t-----\t---------\t---------")
		for _, t := range trains {
			cat := ""
			if t.TrainCategory.Valid {
				cat = t.TrainCategory.String + " "
			}
			avgDelay := 0.0
			if t.AvgDelay.Valid {
				avgDelay = t.AvgDelay.Float64
			}
			maxDelay := 0
			if v, ok := t.MaxDelay.(int64); ok {
				maxDelay = int(v)
			} else if v, ok := t.MaxDelay.(float64); ok {
				maxDelay = int(v)
			}
			fmt.Fprintf(w, "%s%s\t%s → %s\t%d\t%.1f min\t%d min\n",
				cat, t.TrainNumber, t.Origin, t.Destination,
				t.TripCount, avgDelay, maxDelay)
		}
		w.Flush()

	case "reliable":
		trains, err := queries.GetMostReliableTrains(ctx, sqlc.GetMostReliableTrainsParams{
			FromDate:   from,
			ToDate:     to,
			LimitCount: 10,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error querying: %v\n", err)
			os.Exit(1)
		}

		if len(trains) == 0 {
			fmt.Println("No data found")
			return
		}

		fmt.Println("Most reliable trains (last 30 days):")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "Train\tRoute\tTrips\tOn-Time Rate\tAvg Delay")
		fmt.Fprintln(w, "-----\t-----\t-----\t------------\t---------")
		for _, t := range trains {
			cat := ""
			if t.TrainCategory.Valid {
				cat = t.TrainCategory.String + " "
			}
			avgDelay := 0.0
			if t.AvgDelay.Valid {
				avgDelay = t.AvgDelay.Float64
			}
			fmt.Fprintf(w, "%s%s\t%s → %s\t%d\t%.1f%%\t%.1f min\n",
				cat, t.TrainNumber, t.Origin, t.Destination,
				t.TripCount, float64(t.OnTimeRate), avgDelay)
		}
		w.Flush()

	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s (use 'delayed' or 'reliable')\n", subCmd)
		os.Exit(1)
	}
}

