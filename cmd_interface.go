package zfs

const (
	zfsListFilesystems zfsListType = iota
	zfsListSnapshots
)

type zfsListType uint8

var (
	zfsLisTypeToStr = map[zfsListType]string{
		zfsListFilesystems: "filesystem",
		zfsListSnapshots:   "snapshot",
	}
)

type zfsCmd interface {
	list(pool string, recursive bool, listType zfsListType, cols []string) (string, error)
	get(fsOrSnap string, props []string, cols []string) (string, error)
	holds(snap string) (string, error)
}

type zpoolCmd interface {
	list(cols []string) (string, error)
	get(pool string, props []string, cols []string) (string, error)
}

type cmdInvoker struct {
	zfs   zfsCmd
	zpool zpoolCmd
}
