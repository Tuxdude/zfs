package zfs

import (
	"fmt"
	"sort"
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

// RecursiveHoldGroup represents a group of holds with the same tag name applied on a recursive snapshot group.
type RecursiveHoldGroup struct {
	Tag                    string
	Creation               time.Time
	RecursiveSnapshotGroup *RecursiveSnapshotGroup
}

// RecursiveHoldGroupList represents a list of RecursiveHoldGroup objects.
type RecursiveHoldGroupList []*RecursiveHoldGroup

// String returns the string representation of the recursive hold group.
func (r *RecursiveHoldGroup) String() string {
	return fmt.Sprintf(
		"{RecursiveHoldGroup Tag: %q, Creation: %v, RSG: %v}",
		r.Tag,
		r.Creation,
		r.RecursiveSnapshotGroup,
	)
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

func listRecursiveHoldGroups(rsg *RecursiveSnapshotGroup) (RecursiveHoldGroupList, error) {
	// For each snapshot within the rsg, list all the holds.
	var fullFsList []string
	for _, s := range rsg.Snapshots {
		fullFsList = append(fullFsList, s.FileSystem.Name)
	}

	// For each hold tag as the key, build a map with the rsg file system as values.
	holdFsMap := make(map[string][]string)
	holdCtimeMap := make(map[string]time.Time)

	for _, s := range rsg.Snapshots {
		sHolds, err := s.Holds()
		if err != nil {
			return nil, err
		}

		for _, h := range sHolds {
			holdFsMap[h.Tag] = append(holdFsMap[h.Tag], s.FileSystem.Name)

			ctime, ok := holdCtimeMap[h.Tag]
			if ok && !ctime.Equal(h.Creation) {
				return nil, fmt.Errorf("found same hold tag name %q but created at different timestamps, %v %v", h.Tag, ctime, h.Creation)
			}

			holdCtimeMap[h.Tag] = h.Creation
		}
	}

	// Remove keys in the map which do not have all the file systems of the rsg.
	for tag, dsList := range holdFsMap {
		if !strSlicesEqual(fullFsList, dsList) {
			delete(holdFsMap, tag)
		}
	}

	// Remaining keys are the holds of interest.
	// Create a recursive hold group for every such key.
	var result RecursiveHoldGroupList

	for tag := range holdFsMap {
		rhg := &RecursiveHoldGroup{
			Tag:                    tag,
			Creation:               holdCtimeMap[tag],
			RecursiveSnapshotGroup: rsg,
		}
		result = append(result, rhg)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Creation.After(result[j].Creation)
	})

	return result, nil
}
