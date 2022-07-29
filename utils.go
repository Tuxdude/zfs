package zfs

func strSlicesEqual(s1 []string, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	l := len(s1)
	for i := 0; i < l; i++ {
		if s1[i] != s2[i] {
			return false
		}
	}

	return true
}
