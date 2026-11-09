package repository

import "fmt"

// MaxIDLength is the MOS limit for roID, storyID, itemID, objID and mosID.
const MaxIDLength = 128

// CheckID returns ErrInvalid unless id is a usable MOS ID.
func CheckID(kind, id string) error {
	if id == "" {
		return fmt.Errorf("%s ID is required: %w", kind, ErrInvalid)
	}
	if len([]rune(id)) > MaxIDLength {
		return fmt.Errorf("%s ID longer than %d characters: %w", kind, MaxIDLength, ErrInvalid)
	}
	return nil
}

// The ordering helpers below compute a new order of child IDs (stories in a
// running order, items in a story). Both implementations use them, so Move,
// Swap and ReplaceAll behave the same everywhere.

// MoveIDs returns order with ids moved, in the given order, to just before
// the ID before. An empty before moves them to the end. The other IDs keep
// their relative order.
func MoveIDs(order, ids []string, before string) ([]string, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("move needs at least one ID: %w", ErrInvalid)
	}
	if err := CheckDistinct(ids); err != nil {
		return nil, err
	}
	present := make(map[string]bool, len(order))
	for _, id := range order {
		present[id] = true
	}
	moving := make(map[string]bool, len(ids))
	for _, id := range ids {
		if !present[id] {
			return nil, fmt.Errorf("%s: %w", id, ErrNotFound)
		}
		moving[id] = true
	}
	if before != "" {
		if moving[before] {
			return nil, fmt.Errorf("cannot move %s before itself: %w", before, ErrInvalid)
		}
		if !present[before] {
			return nil, fmt.Errorf("%s: %w", before, ErrNotFound)
		}
	}

	out := make([]string, 0, len(order))
	for _, id := range order {
		if moving[id] {
			continue
		}
		if id == before {
			out = append(out, ids...)
		}
		out = append(out, id)
	}
	if before == "" {
		out = append(out, ids...)
	}
	return out, nil
}

// SwapIDs returns order with a and b exchanged.
func SwapIDs(order []string, a, b string) ([]string, error) {
	if a == b {
		return nil, fmt.Errorf("cannot swap %s with itself: %w", a, ErrInvalid)
	}
	ia, ib := -1, -1
	for i, id := range order {
		switch id {
		case a:
			ia = i
		case b:
			ib = i
		}
	}
	if ia < 0 {
		return nil, fmt.Errorf("%s: %w", a, ErrNotFound)
	}
	if ib < 0 {
		return nil, fmt.Errorf("%s: %w", b, ErrNotFound)
	}
	out := append([]string(nil), order...)
	out[ia], out[ib] = out[ib], out[ia]
	return out, nil
}

// CheckDistinct returns ErrInvalid if an ID appears more than once.
func CheckDistinct(ids []string) error {
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			return fmt.Errorf("%s listed twice: %w", id, ErrInvalid)
		}
		seen[id] = true
	}
	return nil
}
