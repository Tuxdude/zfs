package zfs

import (
	"fmt"
	"sort"
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

// Holds returns the list of holds on the snapshot.
func (s *Snapshot) Holds() (HoldList, error) {
	return listHolds(s)
}

// GetProp returns the specified property's value for the snapshot.
func (s *Snapshot) GetProp(prop string) (string, error) {
	return getPropForFsOrSnap(s.FullName(), prop)
}

// RecursiveSnapshotGroup represents a recursive group of snapshots within a pool taken atomically at the same timestamp.
type RecursiveSnapshotGroup struct {
	Name      string
	Creation  time.Time
	Pool      *Pool
	Snapshots SnapshotList
}

// RecursiveSnapshotGroupList represents a list of RecursiveSnapshotGroup objects.
type RecursiveSnapshotGroupList []*RecursiveSnapshotGroup

// String returns the string representation of the recursive snapshot group.
func (r *RecursiveSnapshotGroup) String() string {
	return fmt.Sprintf(
		"{RecursiveSnapshotGroup Name: %q, Creation: %v, Pool: %v}",
		r.Name,
		r.Creation,
		r.Pool,
	)
}

// Holds returns the list of recursive hold groups for the recursive snapshot group.
func (r *RecursiveSnapshotGroup) Holds() (RecursiveHoldGroupList, error) {
	return listRecursiveHoldGroups(r)
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

func listRecursiveSnapshotGroups(pool *Pool) (RecursiveSnapshotGroupList, error) {
	fsList, err := pool.FileSystems()
	if err != nil {
		return nil, err
	}

	var poolFsNames []string
	for _, f := range fsList {
		poolFsNames = append(poolFsNames, f.Name)
	}

	// Hash map with snapshot name as the key and the list of
	// associated snapshots as the value.
	snapMap := make(map[string]SnapshotList)

	// Identify the list of snapshots per file system.
	for _, f := range fsList {
		snapshots, err := f.Snapshots()
		if err != nil {
			return nil, err
		}

		for _, s := range snapshots {
			snapMap[s.Name] = append(snapMap[s.Name], s)
		}
	}

	// Remove all incomplete groups (i.e. a snapshot group that doesn't
	// cover all the file systems within the pool).
	for name, sg := range snapMap {
		var sgFsNames []string
		for _, snap := range sg {
			sgFsNames = append(sgFsNames, snap.FileSystem.Name)
		}

		if !strSlicesEqual(poolFsNames, sgFsNames) {
			delete(snapMap, name)
		}

		// Validate that all snapshots in the group have the same
		// snapshot creation timestamp.
		l := len(sg)
		for i := 1; i < l; i++ {
			if !sg[0].Creation.Equal(sg[i].Creation) {
				return nil, fmt.Errorf("snapshot with the same name has different timestamps across file systems - %v %v", sg[0].FullName(), sg[i].FullName())
			}
		}
	}

	// Convert the map to a list of snapshot groups.
	var sgList []SnapshotList
	for _, sg := range snapMap {
		sgList = append(sgList, sg)
	}

	// Sort the snapshot groups by descending order of the snapshot
	// creation timestamp.
	sort.Slice(sgList, func(i, j int) bool {
		return sgList[i][0].Creation.After(sgList[j][0].Creation)
	})

	// Pack the snapshot group list as list of recursiveSnapshotGroup objects.
	var result RecursiveSnapshotGroupList

	for _, s := range sgList {
		r := &RecursiveSnapshotGroup{
			Name:      s[0].Name,
			Creation:  s[0].Creation,
			Pool:      pool,
			Snapshots: s,
		}
		result = append(result, r)
	}

	return result, nil
}
