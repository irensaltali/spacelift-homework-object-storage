package storage

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	// ErrInvalidObjectID is returned when the object ID does not match input constraints.
	ErrInvalidObjectID = errors.New("invalid object id")
	// ErrObjectNotFound is returned when the object does not exist in storage.
	ErrObjectNotFound = errors.New("object not found")
)

var objectIDPattern = regexp.MustCompile(`^[a-zA-Z0-9]{1,32}$`)

// ValidateObjectID validates the object identifier against API requirements.
func ValidateObjectID(objectID string) error {
	if !objectIDPattern.MatchString(objectID) {
		return fmt.Errorf("%w: expected 1-32 alphanumeric characters", ErrInvalidObjectID)
	}

	return nil
}
