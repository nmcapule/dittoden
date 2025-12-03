package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/lmittmann/tint"
	schema "github.com/nmcapule/dittoden/gen/schema/v1"
	"github.com/nmcapule/dittoden/pkg/registry"
	"github.com/spf13/cobra"
)

var (
	searchName      string
	searchRelatedTo string
	searchProperty  string // format: key=value
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for entities",
	Long:  `Search for entities by name, relation, or property.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelWarn}))

		// Load registry
		records, err := registry.ParseRecordsFromDir(dataDir, logger)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading records: %v\n", err)
			os.Exit(1)
		}

		reg := &registry.Registry{
			Entities:          make(map[string]*schema.Entity),
			Relationships:     make(map[string]*schema.Relationship),
			RelationshipTypes: make(map[string]*schema.RelationshipType),
			Logger:            logger,
		}
		if err := reg.Add(records); err != nil {
			fmt.Fprintf(os.Stderr, "Error adding records to registry: %v\n", err)
			os.Exit(1)
		}

		var results []*registry.SearchResult

		if searchName != "" {
			results = reg.SearchByName(searchName)
		} else if searchRelatedTo != "" {
			results = reg.SearchByRelation(searchRelatedTo)
		} else if searchProperty != "" {
			parts := strings.SplitN(searchProperty, "=", 2)
			if len(parts) != 2 {
				fmt.Fprintln(os.Stderr, "Property search format must be key=value")
				os.Exit(1)
			}
			results = reg.SearchByProperty(parts[0], parts[1])
		} else {
			cmd.Help()
			return
		}

		printResults(results)
	},
}

func printResults(results []*registry.SearchResult) {
	if len(results) == 0 {
		fmt.Println("No results found.")
		return
	}

	fmt.Printf("Found %d results:\n", len(results))
	for _, res := range results {
		fmt.Printf("- [%s] %s\n", res.Entity.Code, getPrimaryLabel(res.Entity))
		fmt.Printf("  Reason: %s\n", res.Reason)
	}
}

func getPrimaryLabel(e *schema.Entity) string {
	for _, l := range e.Labels {
		if l.Type == schema.Entity_Label_LABEL_TYPE_PRIMARY {
			return l.Label
		}
	}
	if len(e.Labels) > 0 {
		return e.Labels[0].Label
	}
	return "(no label)"
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVar(&searchName, "name", "", "Search by partial name matching")
	searchCmd.Flags().StringVar(&searchRelatedTo, "related-to", "", "Search by relation to entity code")
	searchCmd.Flags().StringVar(&searchProperty, "property", "", "Search by property (key=value)")
}
