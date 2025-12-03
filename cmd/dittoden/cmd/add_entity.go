package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	schema "github.com/nmcapule/dittoden/gen/schema/v1"
	"github.com/nmcapule/dittoden/pkg/registry"
	"github.com/spf13/cobra"
)

var (
	addEntityCode  string
	addEntityLabel string
	addEntityType  string
	addEntityFile  string
)

var addEntityCmd = &cobra.Command{
	Use:   "add-entity",
	Short: "Add a new entity to a .txtpb file",
	Long:  `Add a new entity with a code, primary label, and type to a specified .txtpb file.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Gather inputs
		code, err := getCode()
		if err != nil {
			fmt.Printf("Error getting code: %v\n", err)
			os.Exit(1)
		}

		label, err := getLabel()
		if err != nil {
			fmt.Printf("Error getting label: %v\n", err)
			os.Exit(1)
		}

		entityType, err := getType()
		if err != nil {
			fmt.Printf("Error getting type: %v\n", err)
			os.Exit(1)
		}

		filePath, err := getFile()
		if err != nil {
			fmt.Printf("Error getting file: %v\n", err)
			os.Exit(1)
		}

		// 2. Create Entity
		entity := &schema.Entity{
			Code: code,
			Type: entityType,
			Labels: []*schema.Entity_Label{
				{
					Label: label,
					Type:  schema.Entity_Label_LABEL_TYPE_PRIMARY,
				},
			},
		}

		// 3. Read existing records
		records, err := registry.ReadRecordsFromFile(filePath)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", filePath, err)
			os.Exit(1)
		}

		// 4. Check for duplicates (simple check within the file)
		for _, e := range records.Entity {
			if e.Code == code {
				fmt.Printf("Error: Entity with code '%s' already exists in %s\n", code, filePath)
				os.Exit(1)
			}
		}

		// 5. Append and Save
		records.Entity = append(records.Entity, entity)
		if err := registry.WriteRecordsToFile(filePath, records); err != nil {
			fmt.Printf("Error writing to file %s: %v\n", filePath, err)
			os.Exit(1)
		}

		fmt.Printf("Successfully added entity '%s' to %s\n", code, filePath)
	},
}

func init() {
	rootCmd.AddCommand(addEntityCmd)

	addEntityCmd.Flags().StringVar(&addEntityCode, "code", "", "Unique code for the entity")
	addEntityCmd.Flags().StringVar(&addEntityLabel, "label", "", "Primary label for the entity")
	addEntityCmd.Flags().StringVar(&addEntityType, "type", "", "Type of the entity (BEING, ORGANIZATION, LOCATION, EVENT, ARTIFACT)")
	addEntityCmd.Flags().StringVar(&addEntityFile, "file", "", "Path to the .txtpb file")
}

func getCode() (string, error) {
	if addEntityCode != "" {
		return addEntityCode, nil
	}
	prompt := promptui.Prompt{
		Label: "Entity Code",
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("code cannot be empty")
			}
			return nil
		},
	}
	return prompt.Run()
}

func getLabel() (string, error) {
	if addEntityLabel != "" {
		return addEntityLabel, nil
	}
	prompt := promptui.Prompt{
		Label: "Primary Label",
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("label cannot be empty")
			}
			return nil
		},
	}
	return prompt.Run()
}

func getType() (schema.EntityType, error) {
	types := []string{"BEING", "ORGANIZATION", "LOCATION", "EVENT", "ARTIFACT"}
	typeMap := map[string]schema.EntityType{
		"BEING":        schema.EntityType_ENTITY_TYPE_BEING,
		"ORGANIZATION": schema.EntityType_ENTITY_TYPE_ORGANIZATION,
		"LOCATION":     schema.EntityType_ENTITY_TYPE_LOCATION,
		"EVENT":        schema.EntityType_ENTITY_TYPE_EVENT,
		"ARTIFACT":     schema.EntityType_ENTITY_TYPE_ARTIFACT,
	}

	if addEntityType != "" {
		upperType := strings.ToUpper(addEntityType)
		if val, ok := typeMap[upperType]; ok {
			return val, nil
		}
		return schema.EntityType_ENTITY_TYPE_UNSPECIFIED, fmt.Errorf("invalid type: %s", addEntityType)
	}

	prompt := promptui.Select{
		Label: "Entity Type",
		Items: types,
	}

	_, result, err := prompt.Run()
	if err != nil {
		return schema.EntityType_ENTITY_TYPE_UNSPECIFIED, err
	}

	return typeMap[result], nil
}

func getFile() (string, error) {
	if addEntityFile != "" {
		return addEntityFile, nil
	}
	prompt := promptui.Prompt{
		Label: "Target File (.txtpb)",
		Validate: func(input string) error {
			if len(input) == 0 {
				return errors.New("file path cannot be empty")
			}
			if !strings.HasSuffix(input, ".txtpb") {
				return errors.New("file must have .txtpb extension")
			}
			return nil
		},
	}
	return prompt.Run()
}
