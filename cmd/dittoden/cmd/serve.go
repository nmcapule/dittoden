package cmd

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	schema "github.com/nmcapule/dittoden/gen/schema/v1"
	"github.com/nmcapule/dittoden/pkg/registry"
	"github.com/nmcapule/dittoden/pkg/server"
	"github.com/spf13/cobra"
)

var (
	serverPort int
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Dittoden web server",
	Long:  `Start a web server to visualize entities and relationships.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelInfo}))

		// Load registry
		records, err := registry.ParseRecordsFromDir(dataDir, logger)
		if err != nil {
			logger.Error("Failed to parse records from directory", slog.String("dir", dataDir), slog.Any("error", err))
			os.Exit(1)
		}

		reg := &registry.Registry{
			Entities:          make(map[string]*schema.Entity),
			Relationships:     make(map[string]*schema.Relationship),
			RelationshipTypes: make(map[string]*schema.RelationshipType),
			Logger:            logger,
		}
		if err := reg.Add(records); err != nil {
			logger.Error("Failed to add records to registry", slog.Any("error", err))
			os.Exit(1)
		}

		srv := &server.Server{
			Registry: reg,
			Logger:   logger,
			Port:     serverPort,
		}

		if err := srv.Start(); err != nil {
			logger.Error("Server failed", slog.Any("error", err))
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVar(&serverPort, "port", 8080, "Port to run the server on")
}
