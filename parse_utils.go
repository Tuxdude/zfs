package zfs

import (
	"fmt"
	"strconv"
	"strings"
)

func splitOnNewLine(input string) []string {
	s := strings.ReplaceAll(strings.TrimSpace(input), "\r\n", "\n")
	if s == "" {
		return nil
	}

	return strings.Split(s, "\n")
}

func parseUint64(str string, desc string) (uint64, error) {
	res, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing %q, unable to convert %q to uint64: %w", desc, str, err)
	}

	return res, nil
}

func parseUint8(str string, desc string) (uint8, error) {
	res, err := strconv.ParseUint(str, 10, 8)
	if err != nil {
		return 0, fmt.Errorf("parsing %q, unable to convert %q to uint8: %w", desc, str, err)
	}

	return uint8(res), nil
}
