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
	path = flag.String("path", "", "Path to the folder containing .txtpb files to validate.")
)

func main() {
	flag.Parse()

	filepath.Walk(*path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		log.Printf("Visiting: %s", path)
		if !info.IsDir() && filepath.Ext(info.Name()) == ".txtpb" {
			log.Printf("Found file: %s", info.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				log.Printf("Error reading file %s: %v", path, err)
				return err
			}

			entity := &schema.Records{}
			if err := prototext.Unmarshal(data, entity); err != nil {
				log.Printf("Error unmarshalling .txtpb file %s: %v", path, err)
				return err
			}

			log.Printf("Successfully read: %+v", prototext.Format(entity))
		}
		return nil
	})
}
