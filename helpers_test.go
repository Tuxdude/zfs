package zfs

import (
	"fmt"
	"testing"

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

func fakeZpoolWithPropertyOverride(override propMap) *fakeZpool {
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

	for k, v := range override {
		result.props[k] = v
	}
	return result
}

func newPoolForTesting(t *testing.T, poolName string, override propMap) *Pool {
	system := newFakeSystem(
		fakeZpools{
			poolName: fakeZpoolWithPropertyOverride(override),
		})

	pools, gotErr := system.ListPools()
	if nil != gotErr {
		t.Errorf(
			"Test Case: %q\nFailure: gotErr != nil\nReason: %v",
			t.Name(), gotErr)
	}

	if len(pools) != 1 {
		t.Errorf(
			"Test Case: %q\nFailure: Expected exactly 1 pool, but got %d\nReason: pools = %v",
			t.Name(), len(pools), pools)
	}

	return pools[0]
}
