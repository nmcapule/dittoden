package registry

import (
	"fmt"
	"os"

	schema "github.com/nmcapule/dittoden/gen/schema/v1"
	"google.golang.org/protobuf/encoding/prototext"
)

// ReadRecordsFromFile reads a .txtpb file and returns the Records message.
// If the file does not exist, it returns an empty Records message (and no error).
func ReadRecordsFromFile(path string) (*schema.Records, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &schema.Records{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	records := &schema.Records{}
	if err := prototext.Unmarshal(data, records); err != nil {
		return nil, fmt.Errorf("failed to unmarshal records: %w", err)
	}

	return records, nil
}

// WriteRecordsToFile writes the Records message to a .txtpb file.
func WriteRecordsToFile(path string, records *schema.Records) error {
	marshalOpts := prototext.MarshalOptions{
		Multiline: true,
		Indent:    "  ",
	}
	data, err := marshalOpts.Marshal(records)
	if err != nil {
		return fmt.Errorf("failed to marshal records: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
