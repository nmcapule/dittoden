package main

import (
	"flag"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	schema "github.com/nmcapule/dittoden/gen/schema/v1"

	"google.golang.org/protobuf/encoding/prototext"
)

var (
	dir = flag.String("dir", "", "Path to the folder containing .txtpb files to validate.")
)

func main() {
	flag.Parse()

	records := &schema.Records{}
	err := filepath.Walk(*dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		log.Printf("Visiting: %s", path)
		if !info.IsDir() && filepath.Ext(info.Name()) == ".txtpb" {
			data, err := os.ReadFile(path)
			if err != nil {
				log.Printf("Error reading file %s: %v", path, err)
				return err
			}

			filerecords := &schema.Records{}
			if err := prototext.Unmarshal(data, filerecords); err != nil {
				log.Printf("Error unmarshalling .txtpb file %s: %v", path, err)
				return err
			}
			records.Entity = append(records.Entity, filerecords.Entity...)
			records.Relationship = append(records.Relationship, filerecords.Relationship...)

			log.Printf("(%s) Entities: %d, Relationships: %d", path, len(filerecords.Entity), len(filerecords.Relationship))
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error walking the path %s: %v", *dir, err)
	}

	log.Printf("Total Entities: %d, Total Relationships: %d", len(records.Entity), len(records.Relationship))
}
