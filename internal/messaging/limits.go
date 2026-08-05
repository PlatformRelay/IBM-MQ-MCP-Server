package messaging

import "fmt"

// NormalizeBrowseCount clamps requested to [1, MaxBrowseCount], defaulting to DefaultBrowseCount.
func NormalizeBrowseCount(requested int) int {
	if requested <= 0 {
		return DefaultBrowseCount
	}
	if requested > MaxBrowseCount {
		return MaxBrowseCount
	}
	return requested
}

// ValidateBrowseCount returns an error when requested exceeds MaxBrowseCount without clamping.
func ValidateBrowseCount(requested int) error {
	if requested > MaxBrowseCount {
		return fmt.Errorf("count %d exceeds maximum %d", requested, MaxBrowseCount)
	}
	return nil
}

// NormalizeMaxPayloadBytes clamps requested to [1, HardMaxPayloadBytes], defaulting to DefaultMaxPayloadBytes.
func NormalizeMaxPayloadBytes(requested int) int {
	if requested <= 0 {
		return DefaultMaxPayloadBytes
	}
	if requested > HardMaxPayloadBytes {
		return HardMaxPayloadBytes
	}
	return requested
}

// ValidateMaxPayloadBytes returns an error when requested exceeds HardMaxPayloadBytes without clamping.
func ValidateMaxPayloadBytes(requested int) error {
	if requested > HardMaxPayloadBytes {
		return fmt.Errorf("maxPayloadBytes %d exceeds maximum %d", requested, HardMaxPayloadBytes)
	}
	return nil
}
