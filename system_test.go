package zfs

import (
	"regexp"
	"testing"
)

var listPoolsTests = []struct {
	name  string
	pools fakeZpools
	want  PoolList
}{
	{
		name:  "No zpools",
		pools: nil,
	},
	{
		name: "Single pool",
		pools: fakeZpools{
			"TestPool1": &fakeZpool{
				props: propMap{
					"guid":          "123456789012345",
					"size":          "16000000",
					"allocated":     "10000000",
					"free":          "6000000",
					"fragmentation": "5",
					"health":        "ONLINE",
					"altroot":       "-",
				},
			},
		},
		want: PoolList{
			&Pool{
				Name:                 "TestPool1",
				GUID:                 123456789012345,
				Size:                 16000000,
				Allocated:            10000000,
				Free:                 6000000,
				FragmentationPercent: 5,
				HealthStatus:         "ONLINE",
				AltRoot:              "-",
			},
		},
	},
	{
		name: "Six pools with various Health Status and AltRoot",
		pools: fakeZpools{
			"TestPool1": &fakeZpool{
				props: propMap{
					"guid":          "1234567890123451",
					"size":          "18000000",
					"allocated":     "14000000",
					"free":          "4000000",
					"fragmentation": "10",
					"health":        "DEGRADED",
					"altroot":       "-",
				},
			},
			"TestPool2": &fakeZpool{
				props: propMap{
					"guid":          "1234567890123452",
					"size":          "10000000",
					"allocated":     "1000000",
					"free":          "9000000",
					"fragmentation": "11",
					"health":        "FAULTED",
					"altroot":       "/foo-bar",
				},
			},
			"TestPool3": &fakeZpool{
				props: propMap{
					"guid":          "1234567890123453",
					"size":          "20000000",
					"allocated":     "8000000",
					"free":          "12000000",
					"fragmentation": "12",
					"health":        "OFFLINE",
					"altroot":       "/foo-bar-baz",
				},
			},
			"TestPool4": &fakeZpool{
				props: propMap{
					"guid":          "1234567890123454",
					"size":          "40000000",
					"allocated":     "9000000",
					"free":          "31000000",
					"fragmentation": "13",
					"health":        "REMOVED",
					"altroot":       "-",
				},
			},
			"TestPool5": &fakeZpool{
				props: propMap{
					"guid":          "1234567890123455",
					"size":          "12000000",
					"allocated":     "6000000",
					"free":          "6000000",
					"fragmentation": "9",
					"health":        "UNAVAIL",
					"altroot":       "/some-other-root",
				},
			},
			"TestPool6": &fakeZpool{
				props: propMap{
					"guid":          "1234567890123456",
					"size":          "25000000",
					"allocated":     "10000000",
					"free":          "15000000",
					"fragmentation": "1",
					"health":        "ONLINE",
					"altroot":       "-",
				},
			},
		},
		want: PoolList{
			&Pool{
				Name:                 "TestPool1",
				GUID:                 1234567890123451,
				Size:                 18000000,
				Allocated:            14000000,
				Free:                 4000000,
				FragmentationPercent: 10,
				HealthStatus:         "DEGRADED",
				AltRoot:              "-",
			},
			&Pool{
				Name:                 "TestPool2",
				GUID:                 1234567890123452,
				Size:                 10000000,
				Allocated:            1000000,
				Free:                 9000000,
				FragmentationPercent: 11,
				HealthStatus:         "FAULTED",
				AltRoot:              "/foo-bar",
			},
			&Pool{
				Name:                 "TestPool3",
				GUID:                 1234567890123453,
				Size:                 20000000,
				Allocated:            8000000,
				Free:                 12000000,
				FragmentationPercent: 12,
				HealthStatus:         "OFFLINE",
				AltRoot:              "/foo-bar-baz",
			},
			&Pool{
				Name:                 "TestPool4",
				GUID:                 1234567890123454,
				Size:                 40000000,
				Allocated:            9000000,
				Free:                 31000000,
				FragmentationPercent: 13,
				HealthStatus:         "REMOVED",
				AltRoot:              "-",
			},
			&Pool{
				Name:                 "TestPool5",
				GUID:                 1234567890123455,
				Size:                 12000000,
				Allocated:            6000000,
				Free:                 6000000,
				FragmentationPercent: 9,
				HealthStatus:         "UNAVAIL",
				AltRoot:              "/some-other-root",
			},
			&Pool{
				Name:                 "TestPool6",
				GUID:                 1234567890123456,
				Size:                 25000000,
				Allocated:            10000000,
				Free:                 15000000,
				FragmentationPercent: 1,
				HealthStatus:         "ONLINE",
				AltRoot:              "-",
			},
		},
	},
}

func TestListPools(t *testing.T) {
	for _, test := range listPoolsTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			system := newFakeSystem(tc.pools)
			updateExpectedPoolsWithSystem(tc.want, system)

			got, gotErr := system.ListPools()
			if nil != gotErr {
				t.Errorf(
					"ListPools()\nTest Case: %q\nFailure: gotErr != nil\nReason: %v",
					tc.name, gotErr)
				return
			}

			if matchErr := poolListsEqual(tc.want, got); matchErr != nil {
				t.Errorf(
					"ListPools()\nTest Case: %q\nFailure: want and got differ\nReason: %s",
					tc.name, matchErr)
			}
		})
	}
}

var listPoolsErrorTests = []struct {
	name  string
	pools fakeZpools
	want  string
}{
	{
		name: "Invalid pool GUID",
		pools: fakeZpools{
			"TestPool": fakeZpoolWithPropertyOverride("guid", "invalid-guid"),
		},
		want: `parsing "pool info guid", unable to convert "invalid-guid" to uint64:.*`,
	},
	{
		name: "Invalid pool Size",
		pools: fakeZpools{
			"TestPool": fakeZpoolWithPropertyOverride("size", "invalid-size"),
		},
		want: `parsing "pool info size", unable to convert "invalid-size" to uint64:.*`,
	},
	{
		name: "Invalid pool Allocated",
		pools: fakeZpools{
			"TestPool": fakeZpoolWithPropertyOverride("allocated", "invalid-allocated"),
		},
		want: `parsing "pool info allocated", unable to convert "invalid-allocated" to uint64:.*`,
	},
	{
		name: "Invalid pool Free",
		pools: fakeZpools{
			"TestPool": fakeZpoolWithPropertyOverride("free", "invalid-free"),
		},
		want: `parsing "pool info free", unable to convert "invalid-free" to uint64:.*`,
	},
	{
		name: "Invalid pool Fragmentation",
		pools: fakeZpools{
			"TestPool": fakeZpoolWithPropertyOverride("fragmentation", "invalid-fragmentation"),
		},
		want: `parsing "pool info fragmentation", unable to convert "invalid-fragmentation" to uint8:.*`,
	},
	{
		name: "Invalid pool Health",
		pools: fakeZpools{
			"TestPool": fakeZpoolWithPropertyOverride("health", ""),
		},
		want: `parsing "pool info health", invalid empty health: ""`,
	},
	{
		name: "Invalid pool AltRoot",
		pools: fakeZpools{
			"TestPool": fakeZpoolWithPropertyOverride("altroot", "\r"),
		},
		want: `parsing "pool info altroot", invalid empty altroot: ""`,
	},
}

func TestListPoolsErrors(t *testing.T) {
	for _, test := range listPoolsErrorTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			system := newFakeSystem(tc.pools)

			_, gotErr := system.ListPools()
			if gotErr == nil {
				t.Errorf(
					"ListPools()\nTest Case: %q\nFailure: gotErr == nil\nReason: want = %q",
					tc.name, tc.want)
				return
			}

			match, err := regexp.MatchString(tc.want, gotErr.Error())
			if err != nil {
				t.Errorf(
					"ListPools()\nTest Case: %q\nFailure: unexpected exception while matching against gotErr error string\nReason: error = %v",
					tc.name, err)
				return
			}

			if !match {
				t.Errorf(
					"ListPools()\nTest Case: %q\nFailure: gotErr did not match the want regex\nReason:\n\tgotErr = %q\n\twant = %q",
					tc.name, gotErr, tc.want)
			}
		})
	}
}
