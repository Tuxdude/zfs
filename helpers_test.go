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
