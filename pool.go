package zfs

import (
	"fmt"
	"strings"
)

var (
	listPoolsOutputCols = []string{
		"name",
		"guid",
		"size",
		"allocated",
		"free",
		"fragmentation",
		"health",
		"altroot",
	}
	getPoolPropOutputCols = []string{
		"value",
	}
)

// Pool represents a zpool.
type Pool struct {
	// Name of the pool.
	Name string
	// GUID of the pool.
	GUID uint64
	// Total size of the pool in bytes.
	Size uint64
	// Number of bytes allocated in the pool.
	Allocated uint64
	// Number of bytes free in the pool.
	Free uint64
	// Fragmentation percentage.
	FragmentationPercent uint8
	// Health status of the pool.
	HealthStatus string
	// Alternate root for the pool.
	AltRoot string
}

// PoolList represents a list of Pool objects.
type PoolList []*Pool

// String returns the string representation of the pool.
func (p *Pool) String() string {
	return fmt.Sprintf("{Pool Name: %q}", p.Name)
}

// VerboseString returns a verbose string representation of the pool.
func (p *Pool) VerboseString() string {
	return fmt.Sprintf(
		"{Pool Name: %q, GUID: %d, Size: %d, Allocated: %d, Free: %d, Fragmentation: %d%%, "+
			"HealthStatus: %q, AltRoot: %q}",
		p.Name,
		p.GUID,
		p.Size,
		p.Allocated,
		p.Free,
		p.FragmentationPercent,
		p.HealthStatus,
		p.AltRoot,
	)
}

// FileSystems returns the list of file systems within the pool.
func (p *Pool) FileSystems() (FileSystemList, error) {
	return listFileSystems(p)
}

// RecursiveSnapshotGroups returns the list of groups of recursive snapshots taken atomically at the same timestamp within the pool.
func (p *Pool) RecursiveSnapshotGroups() (RecursiveSnapshotGroupList, error) {
	return listRecursiveSnapshotGroups(p)
}

// GetProp returns the specified property's value for the pool.
func (p *Pool) GetProp(prop string) (string, error) {
	return getPropForPool(p.Name, prop)
}

func parsePoolInfo(line string) (*Pool, error) {
	cols := strings.Split(line, "\t")
	if len(cols) != 8 {
		return nil, fmt.Errorf("expected 8 columns per line in pool info, but found %d, line: %q", len(cols), line)
	}

	guid, err := parseUint64(cols[1], "pool info guid")
	if err != nil {
		return nil, err
	}

	size, err := parseUint64(cols[2], "pool info size")
	if err != nil {
		return nil, err
	}

	alloc, err := parseUint64(cols[3], "pool info allocated")
	if err != nil {
		return nil, err
	}

	free, err := parseUint64(cols[4], "pool info free")
	if err != nil {
		return nil, err
	}

	frag, err := parseUint8(cols[5], "pool info fragmentation")
	if err != nil {
		return nil, err
	}

	return &Pool{
		Name:                 cols[0],
		GUID:                 guid,
		Size:                 size,
		Allocated:            alloc,
		Free:                 free,
		FragmentationPercent: frag,
		HealthStatus:         cols[6],
		AltRoot:              cols[7],
	}, nil
}

// ListPools scans the system for zpools and returns the list of pools found.
func ListPools() (PoolList, error) {
	cmd := systemCmdInvoker()

	out, err := cmd.zpool.list(listPoolsOutputCols)
	if err != nil {
		return nil, fmt.Errorf("failed to list pools, reason: %w", err)
	}

	var result PoolList

	pools := splitOnNewLine(out)
	for _, p := range pools {
		pool, err := parsePoolInfo(p)
		if err != nil {
			return nil, err
		}

		result = append(result, pool)
	}

	return result, nil
}

func getPropForPool(pool string, prop string) (string, error) {
	cmd := systemCmdInvoker()

	out, err := cmd.zpool.get(pool, []string{prop}, getPoolPropOutputCols)
	if err != nil {
		return "", fmt.Errorf(
			"failed to get property %q of pool %q, reason: %w", prop, pool, err)
	}

	val, err := strFromOnlyLine(out)
	if err != nil {
		return "", fmt.Errorf("failed to parse property value %q, reason: %w", out, err)
	}

	return val, nil
}
