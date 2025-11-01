package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/lmittmann/tint"

	schema "github.com/nmcapule/dittoden/gen/schema/v1"

	"google.golang.org/protobuf/encoding/prototext"
)

var (
	dir = flag.String("dir", "", "Path to the folder containing .txtpb files to validate.")
)

type Validator struct {
	Entities      map[string]*schema.Entity
	Relationships map[string]*schema.Relationship
	Logger        *slog.Logger
}

func (v *Validator) RegisterEntity(e *schema.Entity) error {
	// Check for duplicate entity codes.
	if _, exists := v.Entities[e.GetCode()]; exists {
		v.Logger.Error(
			"Duplicate entity found", slog.String("code", e.GetCode()),
			slog.Any("entity", e), slog.Any("existing_entity", v.Entities[e.GetCode()]))
		return fmt.Errorf("duplicate entity found for code: %s", e.GetCode())
	}

	v.Entities[e.GetCode()] = e
	return nil
}

func (v *Validator) RegisterRelationship(r *schema.Relationship) error {
	// Check for duplicate relationship codes.
	if _, exists := v.Relationships[r.GetCode()]; exists {
		v.Logger.Error(
			"Duplicate relationship found", slog.String("code", r.GetCode()),
			slog.Any("relationship", r), slog.Any("existing_relationship", v.Relationships[r.GetCode()]))
		return fmt.Errorf("duplicate relationship found: %s", r.GetCode())
	}

	v.Relationships[r.GetCode()] = r
	return nil
}

// Run validator on all registered entities and relationships.
func (v *Validator) Run() error {
	var errs []error

	// Validate relationships
	for _, r := range v.Relationships {
		a, z := r.GetA(), r.GetZ()

		// Check if source entity exists.
		if _, exists := v.Entities[a.GetCode()]; !exists {
			v.Logger.Error(
				"Non-existent `a` entity code",
				slog.String("relationship", r.GetCode()), slog.String("a_entity", a.GetCode()))
			errs = append(errs, fmt.Errorf("relationship %s has non-existent `a` entity code: %s", r.GetCode(), a.GetCode()))
		}

		// Check if target entity exists.
		if _, exists := v.Entities[z.GetCode()]; !exists {
			v.Logger.Error(
				"Non-existent `z` entity code",
				slog.String("relationship", r.GetCode()), slog.String("z_entity", z.GetCode()))
			errs = append(errs, fmt.Errorf("relationship %s has non-existent `z` entity code: %s", r.GetCode(), z.GetCode()))
		}
	}

	return errors.Join(errs...)
}

func main() {
	flag.Parse()

	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug}))

	validator := &Validator{
		Entities:      make(map[string]*schema.Entity),
		Relationships: make(map[string]*schema.Relationship),
		Logger:        logger,
	}
	err := filepath.Walk(*dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		logger.Debug("Visiting file", slog.String("file", path))
		if !info.IsDir() && filepath.Ext(info.Name()) == ".txtpb" {
			data, err := os.ReadFile(path)
			if err != nil {
				validator.Logger.Error("Error reading file", slog.String("file", path), slog.Any("error", err))
				return err
			}

			filerecords := &schema.Records{}
			if err := prototext.Unmarshal(data, filerecords); err != nil {
				validator.Logger.Error("Error unmarshalling .txtpb file", slog.String("file", path), slog.Any("error", err))
				return err
			}
			for _, e := range filerecords.Entity {
				if err := validator.RegisterEntity(e); err != nil {
					return err
				}
			}
			for _, r := range filerecords.Relationship {
				if err := validator.RegisterRelationship(r); err != nil {
					return err
				}
			}

			validator.Logger.Info(
				"Registered records from file", slog.String("file", path),
				slog.Int("entities", len(filerecords.Entity)), slog.Int("relationships", len(filerecords.Relationship)))
		}
		return nil
	})
	if err != nil {
		validator.Logger.Error("Error walking the path", slog.String("path", *dir), slog.Any("error", err))
		return
	}

	if err := validator.Run(); err != nil {
		validator.Logger.Error("Validator failed", slog.Any("error", err))
		return
	}

	validator.Logger.Info("Total registered records",
		slog.Int("entities", len(validator.Entities)),
		slog.Int("relationships", len(validator.Relationships)))
	validator.Logger.Info("Validation successful")
}
