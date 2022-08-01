package zfs

import (
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func newFakeSystem(pools fakeZpools) *System {
	return NewSystem(&SystemConfig{
		alternateCmd: newFakeZpoolCmd(pools),
	})
}

func getFakeZpoolCmd(system *System) *fakeZpoolCmd {
	result, ok := system.cmd.zpool.(*fakeZpoolCmd)
	if !ok {
		panic("supplied system argument includes a cmd.zpool field that is not a *fakeZpoolCmd")
	}

	return result
}

func updateExpectedPoolsWithSystem(pools PoolList, system *System) {
	for _, pool := range pools {
		pool.System = system
	}
}

func poolListsEqual(actual PoolList, expected PoolList) error {
	if diff := cmp.Diff(actual, expected, cmpopts.IgnoreUnexported(System{})); diff != "" {
		return fmt.Errorf("pool lists are not equal\ndiff:\n%s", diff)
	}

	return nil
}

func fakeZpoolWithPropertyOverride(prop string, val string) *fakeZpool {
	result := &fakeZpool{
		props: propMap{
			"guid":          "1234567890123459",
			"size":          "23000000",
			"allocated":     "19000000",
			"free":          "4000000",
			"fragmentation": "4",
			"health":        "ONLINE",
			"altroot":       "/some-alt-root",
		},
	}
	result.props[prop] = val
	return result
}
