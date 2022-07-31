package zfs

type System struct {
	cmd *cmd
}

type SystemConfig struct {
	alternateCmd *cmd
}

func (s *System) ListPools() (PoolList, error) {
	return listPools(s)
}

func NewSystem(config *SystemConfig) *System {
	result := &System{
		cmd: config.alternateCmd,
	}
	if result.cmd == nil {
		result.cmd = realSystemCmd()
	}

	return result
}
