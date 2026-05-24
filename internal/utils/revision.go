package utils

import (
	"time"
)

// FreeRevisionWindowHours defines the free revision period after CV generation
const FreeRevisionWindowHours = 48

// IsWithinFreeRevisionWindow checks if the CV was created within the free revision window
func IsWithinFreeRevisionWindow(createdAt time.Time) bool {
	return time.Since(createdAt) <= time.Duration(FreeRevisionWindowHours)*time.Hour
}

// GetRevisionDeadline returns the deadline for free revisions
func GetRevisionDeadline(createdAt time.Time) time.Time {
	return createdAt.Add(time.Duration(FreeRevisionWindowHours) * time.Hour)
}

// TimeUntilRevisionDeadline returns the remaining time for free revisions
func TimeUntilRevisionDeadline(createdAt time.Time) time.Duration {
	deadline := GetRevisionDeadline(createdAt)
	remaining := time.Until(deadline)
	if remaining < 0 {
		return 0
	}
	return remaining
}
