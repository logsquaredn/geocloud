package main

import "strconv"

func coalesceString(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

func parseBool(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}

func parseInt64(s string) int64 {
	if i, err := strconv.Atoi(s); err == nil {
		return int64(i)
	}

	return 0
}
