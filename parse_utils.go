package zfs

import (
	"fmt"
	"strconv"
	"strings"
	"time"
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

func parseUnixTimestamp(str string, desc string) (time.Time, error) {
	t, err := parseUint64(str, desc)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(int64(t), 0), nil
}

func parseTimestamp(str string, layout string, desc string) (time.Time, error) {
	loc, err := time.LoadLocation("Local")
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to load local time location")
	}

	t, err := time.ParseInLocation(layout, str, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf(
			"parsing %q, unable to convert %q to timestamp: %w", desc, str, err)
	}

	return t, nil
}
