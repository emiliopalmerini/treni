package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/emiliopalmerini/treni/internal/app"
	"github.com/emiliopalmerini/treni/internal/database"
	"github.com/emiliopalmerini/treni/internal/staticdata"
	staticdataPersistence "github.com/emiliopalmerini/treni/internal/staticdata/persistence"
	stationPersistence "github.com/emiliopalmerini/treni/internal/station/persistence"
	"github.com/emiliopalmerini/treni/internal/viaggiatreno"
)

func main() {
	force := flag.Bool("force", false, "Force import even if data is fresh")
	status := flag.Bool("status", false, "Show import status only")
	flag.Parse()

	cfg, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	metadataRepo := staticdataPersistence.NewSQLiteMetadataRepository(db)

	if *status {
		showStatus(metadataRepo)
		return
	}

	httpClient := viaggiatreno.NewHTTPClient()
	stationRepo := stationPersistence.NewSQLiteRepository(db)
	stationRepoAdapter := staticdata.NewStationRepositoryAdapter(stationRepo)

	scheduler := staticdata.NewImportScheduler(
		metadataRepo,
		httpClient,
		stationRepoAdapter,
		cfg.ImportRefreshInterval,
		cfg.StationStalenessAge,
	)

	ctx := context.Background()

	if *force {
		log.Println("forcing station import...")
		scheduler.ForceRefresh(ctx)
	} else {
		shouldRefresh, err := metadataRepo.ShouldRefresh(ctx, "stations", cfg.StationStalenessAge)
		if err != nil && err != staticdata.ErrNotFound {
			log.Fatalf("failed to check refresh status: %v", err)
		}

		if !shouldRefresh {
			log.Println("station data is still fresh, use -force to reimport")
			showStatus(metadataRepo)
			return
		}

		log.Println("starting station import...")
		scheduler.ForceRefresh(ctx)
	}

	showStatus(metadataRepo)
}

func showStatus(repo staticdata.ImportMetadataRepository) {
	meta, err := repo.Get(context.Background(), "stations")
	if err != nil {
		if err == staticdata.ErrNotFound {
			fmt.Println("No import metadata found. Run import first.")
			return
		}
		log.Printf("failed to get import status: %v", err)
		return
	}

	fmt.Println("\n=== Import Status ===")
	fmt.Printf("Last import:    %s\n", meta.LastImport.Format(time.RFC3339))
	fmt.Printf("Records:        %d\n", meta.RecordCount)
	fmt.Printf("Duration:       %dms\n", meta.DurationMs)
	fmt.Printf("Status:         %s\n", meta.Status)
	if meta.ErrorMessage != "" {
		fmt.Printf("Error:          %s\n", meta.ErrorMessage)
	}
	fmt.Printf("Age:            %s\n", time.Since(meta.LastImport).Round(time.Minute))
}
