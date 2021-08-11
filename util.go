package atomicredteam

import "strings"

func ExpandStringSlice(s []string) []string {
	if len(s) == 0 {
		return nil
	}

	var r []string

	for _, e := range s {
		t := strings.Split(e, ",")
		r = append(r, t...)
	}

	return r
}
