package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"

	schema "github.com/nmcapule/dittoden/gen/schema/v1"
	"github.com/nmcapule/dittoden/pkg/registry"
)

var (
	dir = flag.String("dir", "", "Path to the folder containing .txtpb files to validate.")
)

func main() {
	flag.Parse()

	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug}))

	records, err := registry.ParseRecordsFromDir(*dir, logger)
	if err != nil {
		logger.Error("Failed to parse records from directory", slog.String("dir", *dir), slog.Any("error", err))
		return
	}

	r := &registry.Registry{
		Entities:          make(map[string]*schema.Entity),
		Relationships:     make(map[string]*schema.Relationship),
		RelationshipTypes: make(map[string]*schema.RelationshipType),
		Logger:            logger,
	}
	if err := r.Add(records); err != nil {
		logger.Error("Failed to add records to registry", slog.Any("error", err))
		return
	}
	if err := r.Validate(); err != nil {
		logger.Error("Validator failed", slog.Any("error", err))
		return
	}

	logger.Info("Total registered records",
		slog.Int("entities", len(r.Entities)),
		slog.Int("relationship_types", len(r.RelationshipTypes)),
		slog.Int("relationships", len(r.Relationships)),
	)
	logger.Info("Validation successful")
}
