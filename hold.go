package zfs

import (
	"fmt"
	"strings"
	"time"
)

const (
	holdTimestampLayout = "Mon Jan _2 15:04 2006"
)

// Hold represents a hold on a snapshot of a file system within a zpool.
type Hold struct {
	// Tag name of the hold.
	Tag string
	// Creation time of the hold.
	Creation time.Time
	// Associated snapshot for the hold.
	Snapshot *Snapshot
}

// HoldList represents a list of Hold objects.
type HoldList []*Hold

// String returns the string representation of the hold.
func (h *Hold) String() string {
	return fmt.Sprintf("{Hold Tag: %q, Snapshot: %v}", h.Tag, h.Snapshot)
}

// VerboseString returns a verbose string representation of the hold.
func (h *Hold) VerboseString() string {
	return fmt.Sprintf("{Hold Tag: %q, Creation: %v, Snapshot: %v}", h.Tag, h.Creation, h.Snapshot)
}

func listHolds(snapshot *Snapshot) (HoldList, error) {
	out, err := runZfsCmd("holds", "-H", snapshot.FullName())
	if err != nil {
		return nil, fmt.Errorf("failed to list holds of snapshot %q, reason: %w", snapshot, err)
	}

	var result HoldList

	for _, line := range splitOnNewLine(out) {
		h, err := parseHoldInfo(snapshot, line)
		if err != nil {
			return nil, err
		}

		result = append(result, h)
	}

	return result, nil
}

func parseHoldInfo(snapshot *Snapshot, line string) (*Hold, error) {
	cols := strings.Split(line, "\t")
	if len(cols) != 3 {
		return nil, fmt.Errorf("expected 3 columns per line in snapshot info, but found %d, line: %q", len(cols), line)
	}

	creation, err := parseTimestamp(cols[2], holdTimestampLayout, "hold info creation")
	if err != nil {
		return nil, err
	}

	return &Hold{
		Tag:      cols[1],
		Creation: creation,
		Snapshot: snapshot,
	}, nil
}
