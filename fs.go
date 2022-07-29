package zfs

import (
	"fmt"
	"strings"
	"time"
)

// FileSystem represents a file system within a zpool.
type FileSystem struct {
	// Name of the file system.
	Name string
	// True if this is the root file system within the pool, false otherwise.
	IsRoot bool
	// The parent zpool.
	Pool *Pool
	// GUID of the file system.
	GUID uint64
	// Creation time of the file system.
	Creation time.Time
}

// FileSystemList represents a list of FileSystem objects.
type FileSystemList []*FileSystem

// String returns the string representation of the file system.
func (f *FileSystem) String() string {
	return fmt.Sprintf("{FileSystem Name: %q, IsRoot: %t, Pool: %v}", f.Name, f.IsRoot, f.Pool)
}

// VerboseString returns a verbose string representation of the file system.
func (f *FileSystem) VerboseString() string {
	return fmt.Sprintf(
		"{FileSystem Name: %q, IsRoot: %t, Pool: %v, GUID: %d, Creation: %v}",
		f.Name,
		f.IsRoot,
		f.Pool,
		f.GUID,
		f.Creation,
	)
}

// FullName returns the full path of the file system prefixed by the pool name.
func (f *FileSystem) FullName() string {
	if f.IsRoot {
		return f.Pool.Name
	}

	return fmt.Sprintf("%s/%s", f.Pool.Name, f.Name)
}

// Snapshots returns the list of snapshots associated with this file system.
func (f *FileSystem) Snapshots() (SnapshotList, error) {
	return listSnapshots(f)
}

// GetProp returns the specified property's value for the file system.
func (f *FileSystem) GetProp(prop string) (string, error) {
	return getPropForFsOrSnap(f.FullName(), prop)
}

func listFileSystems(pool *Pool) (FileSystemList, error) {
	out, err := runZfsCmd(
		"list",
		"-H",
		"-p",
		"-r",
		"-t",
		"filesystem",
		"-o",
		"name,guid,creation",
		pool.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to list file systems of %q, reason: %w", pool, err)
	}

	var result FileSystemList

	for _, line := range splitOnNewLine(out) {
		fs, err := parseFileSystemInfo(pool, line)
		if err != nil {
			return nil, err
		}

		result = append(result, fs)
	}

	// TODO: Assert that there are no duplicate file system names, which
	// implicitly also guarantees that there is exactly one file system with
	// IsRoot == true.

	return result, nil
}

func parseFileSystemInfo(pool *Pool, line string) (*FileSystem, error) {
	cols := strings.Split(line, "\t")
	if len(cols) != 3 {
		return nil, fmt.Errorf("expected 3 columns per line in file system info, but found %d, line: %q", len(cols), line)
	}

	name := cols[0]

	guid, err := parseUint64(cols[1], "file system info guid")
	if err != nil {
		return nil, err
	}

	creation, err := parseUnixTimestamp(cols[2], "file system info creation")
	if err != nil {
		return nil, err
	}

	return &FileSystem{
		Name:     strings.TrimPrefix(name, fmt.Sprintf("%s/", pool.Name)),
		IsRoot:   name == pool.Name,
		Pool:     pool,
		GUID:     guid,
		Creation: creation,
	}, nil
}

func getPropForFsOrSnap(fsOrSnap string, prop string) (string, error) {
	out, err := runZfsCmd("get", "-H", "-o", "value", prop, fsOrSnap)
	if err != nil {
		return "", fmt.Errorf(
			"failed to get property %q of filesystem/snapshot %q, reason: %w", prop, fsOrSnap, err)
	}

	val, err := strFromOnlyLine(out)
	if err != nil {
		return "", fmt.Errorf("failed to parse property value %q, reason: %w", out, err)
	}

	return val, nil
}
