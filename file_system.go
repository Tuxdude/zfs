package zfs

import (
	"fmt"
	"strings"
	"time"
)

// FileSystem represents a file system within a zpool.
type FileSystem struct {
	// Name of the filesystem.
	Name string
	// True if this is the root filesystem within the pool, false otherwise.
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
func (d *FileSystem) String() string {
	return fmt.Sprintf("{FileSystem Name: %q, IsRoot: %t, Pool: %v}", d.Name, d.IsRoot, d.Pool)
}

// VerboseString returns a verbose string representation of the file system.
func (d *FileSystem) VerboseString() string {
	return fmt.Sprintf(
		"{FileSystem Name: %q, IsRoot: %t, Pool: %v, GUID: %d, Creation: %v}",
		d.Name,
		d.IsRoot,
		d.Pool,
		d.GUID,
		d.Creation,
	)
}

// FullName returns the full path of the file system prefixed by the pool name.
func (d *FileSystem) FullName() string {
	if d.IsRoot {
		return d.Pool.Name
	}

	return fmt.Sprintf("%s/%s", d.Pool.Name, d.Name)
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
		d, err := parseFileSystemInfo(pool, line)
		if err != nil {
			return nil, err
		}

		result = append(result, d)
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
