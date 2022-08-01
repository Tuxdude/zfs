package zfs

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type propMap map[string]string

type fakeZpools map[string]*fakeZpool

func (f fakeZpools) sortedKeys() []string {
	result := make([]string, len(f))
	i := 0

	for k := range f {
		result[i] = k
		i++
	}

	sort.Strings(result)

	return result
}

type fakeZpoolFileSystems map[string]*fakeZpoolFileSystem

type fakeZpoolSnapshots map[string]*fakeZpoolSnapshot

type fakeZpoolHolds map[string]*fakeZpoolHold

type fakeZpool struct {
	props propMap
	fs    fakeZpoolFileSystems
}

type fakeZpoolFileSystem struct {
	props propMap
	snaps fakeZpoolSnapshots
}

type fakeZpoolSnapshot struct {
	props propMap
	holds fakeZpoolHolds
}

type fakeZpoolHold struct {
	creation time.Time
}

func newFakeZpoolCmd(pools fakeZpools) *cmd {
	updateNameProp(pools)

	return &cmd{
		zpool: &fakeZpoolCmd{pools: pools},
		zfs:   &fakeZfsCmd{pools: pools},
	}
}

func updateNameProp(pools fakeZpools) {
	for poolName, pool := range pools {
		if val, ok := pool.props["name"]; ok {
			panic(fmt.Errorf(
				"pool name %q includes a property \"name=%s\" that is disallowed",
				poolName, val))
		}

		pool.props["name"] = poolName

		for fsName, fs := range pool.fs {
			if val, ok := fs.props["name"]; ok {
				panic(fmt.Errorf(
					"fs name \"%s/%s\" includes a property \"name=%s\" that is disallowed",
					poolName, fsName, val))
			}

			fs.props["name"] = fsName

			for snapName, snap := range fs.snaps {
				if val, ok := snap.props["name"]; ok {
					panic(fmt.Errorf(
						"snap name \"%s/%s@%s\" includes a property \"name=%s\" that is disallowed",
						poolName, fsName, snapName, val))
				}

				snap.props["name"] = snapName

				// TODO: Remove after fake holds are implemented.
				for _, hold := range snap.holds {
					_ = hold.creation
				}
			}
		}
	}
}

type fakeZfsCmd struct {
	pools fakeZpools
}

func (f *fakeZfsCmd) list(fsOrSnap string, recursive bool, listType zfsListType, cols []string) (string, error) {
	if len(cols) == 0 {
		return "", fmt.Errorf("at least one column must be specified for 'zfs list'")
	}

	// TODO: Implement this.
	panic(fmt.Errorf("Unimplemented"))
}

func (f *fakeZfsCmd) get(fsOrSnap string, props []string, cols []string) (string, error) {
	if len(cols) == 0 {
		return "", fmt.Errorf("at least one column must be specified for 'zfs get'")
	}

	// TODO: Implement this.
	panic(fmt.Errorf("Unimplemented"))
}

func (f *fakeZfsCmd) holds(snap string) (string, error) {
	// TODO: Implement this.
	panic(fmt.Errorf("Unimplemented"))
}

type fakeZpoolCmd struct {
	pools fakeZpools
}

func (f *fakeZpoolCmd) list(cols []string) (string, error) {
	if len(cols) == 0 {
		return "", fmt.Errorf("at least one column must be specified for 'zpool list'")
	}

	pools := f.pools.sortedKeys()
	ow := newColOutputWriter()

	for _, pool := range pools {
		ow.writePropertyMap(f.pools[pool].props, cols)
	}

	return ow.String(), nil
}

func (f *fakeZpoolCmd) get(pool string, props []string, cols []string) (string, error) {
	if len(cols) == 0 {
		return "", fmt.Errorf("at least one column must be specified for 'zpool get'")
	}

	// TODO: Implement this.
	panic(fmt.Errorf("Unimplemented"))
}

type colOutputWriter struct {
	result           *strings.Builder
	containsFirstCol bool
}

func (o *colOutputWriter) writePropertyMap(props map[string]string, cols []string) {
	for _, col := range cols {
		o.writeColf("%s", props[col])
	}
	o.writeNewLine()
}

func (o *colOutputWriter) writeColf(format string, args ...interface{}) {
	if o.containsFirstCol {
		fmt.Fprint(o.result, "\t")
	} else {
		o.containsFirstCol = true
	}

	fmt.Fprintf(o.result, format, args...)
}

func (o *colOutputWriter) writeNewLine() {
	fmt.Fprint(o.result, "\n")
	o.containsFirstCol = false
}

func (o *colOutputWriter) String() string {
	return o.result.String()
}

func newColOutputWriter() *colOutputWriter {
	return &colOutputWriter{
		result:           new(strings.Builder),
		containsFirstCol: false,
	}
}
