package zfs

import (
	"testing"
)

var validListPoolsTests = []struct {
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
	for _, test := range validListPoolsTests {
		tc := test
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			system := newFakeSystem(tc.pools)
			updateExpectedPoolsWithSystem(tc.want, system)

			got, gotErr := system.ListPools()
			if nil != gotErr {
				t.Errorf(
					"ListPools()\nTest Case: %q\nFailure: gotErr != nil\nReason: %v",
					tc.name,
					gotErr,
				)
			} else if matchErr := poolListsEqual(tc.want, got); matchErr != nil {
				t.Errorf(
					"ListPools()\nTest Case: %q\nFailure: want and got differ\nReason: %s",
					tc.name,
					matchErr,
				)
			}
		})
	}
}
