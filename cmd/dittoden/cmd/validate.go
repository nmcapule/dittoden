package cmd

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	schema "github.com/nmcapule/dittoden/gen/schema/v1"
	"github.com/nmcapule/dittoden/pkg/registry"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate all .txtpb files in the directory",
	Long:  `Validate all .txtpb files in the directory specified by the --dir flag.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug}))

		records, err := registry.ParseRecordsFromDir(dataDir, logger)
		if err != nil {
			logger.Error("Failed to parse records from directory", slog.String("dir", dataDir), slog.Any("error", err))
			os.Exit(1)
		}

		r := &registry.Registry{
			Entities:          make(map[string]*schema.Entity),
			Relationships:     make(map[string]*schema.Relationship),
			RelationshipTypes: make(map[string]*schema.RelationshipType),
			Logger:            logger,
		}
		if err := r.Add(records); err != nil {
			logger.Error("Failed to add records to registry", slog.Any("error", err))
			os.Exit(1)
		}
		if err := r.Validate(); err != nil {
			logger.Error("Validator failed", slog.Any("error", err))
			os.Exit(1)
		}

		logger.Info("Total registered records",
			slog.Int("entities", len(r.Entities)),
			slog.Int("relationship_types", len(r.RelationshipTypes)),
			slog.Int("relationships", len(r.Relationships)),
		)
		logger.Info("Validation successful")
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
