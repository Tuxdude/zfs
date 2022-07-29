package zfs

import (
	"fmt"
	"strings"
	"time"
)

// Snapshot represents a snapshot of the file system within a zpool.
type Snapshot struct {
	// Name of the snapshot.
	Name string
	// Associated file system for the snapshot.
	FileSystem *FileSystem
	// GUID of the snapshot.
	GUID uint64
	// Creation time of the snapshot.
	Creation time.Time
}

// SnapshotList represents a list of Snapshot objects.
type SnapshotList []*Snapshot

// String returns the string representation of the snapshot.
func (s *Snapshot) String() string {
	return fmt.Sprintf("{Snapshot Name: %q, FileSystem: %v}", s.Name, s.FileSystem)
}

// VerboseString returns a verbose string representation of the snapshot.
func (s *Snapshot) VerboseString() string {
	return fmt.Sprintf(
		"{Snapshot Name: %q, FileSystem: %v, Guid: %d, Creation: %v}",
		s.Name,
		s.FileSystem,
		s.GUID,
		s.Creation,
	)
}

// FullName returns the snapshot name prefixed by '@', in turn prefixed by the full path of the file system, in turn prefixed by the pool name.
func (s *Snapshot) FullName() string {
	return fmt.Sprintf("%s@%s", s.FileSystem.FullName(), s.Name)
}

func listSnapshots(fs *FileSystem) (SnapshotList, error) {
	out, err := runZfsCmd("list", "-H", "-p", "-t", "snapshot", "-o", "name,guid,creation", fs.FullName())
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots of file system %q, reason: %w", fs, err)
	}

	var result SnapshotList

	for _, line := range splitOnNewLine(out) {
		s, err := parseSnapshotInfo(fs, line)
		if err != nil {
			return nil, err
		}

		result = append(result, s)
	}

	return result, nil
}

func parseSnapshotInfo(fs *FileSystem, line string) (*Snapshot, error) {
	cols := strings.Split(line, "\t")
	if len(cols) != 3 {
		return nil, fmt.Errorf("expected 3 columns per line in snapshot info, but found %d, line: %q", len(cols), line)
	}

	name := cols[0]

	guid, err := parseUint64(cols[1], "snapshot info guid")
	if err != nil {
		return nil, err
	}

	creation, err := parseUnixTimestamp(cols[2], "snapshot info creation")
	if err != nil {
		return nil, err
	}

	return &Snapshot{
		Name:       strings.TrimPrefix(name, fmt.Sprintf("%s@", fs.FullName())),
		FileSystem: fs,
		GUID:       guid,
		Creation:   creation,
	}, nil
}
