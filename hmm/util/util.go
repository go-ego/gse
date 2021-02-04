package util

import "regexp"

/*
RegexpSplit split slices s into substrings separated by the expression and
returns a slice

This function acts consistent with Python's re.split function.
*/
func RegexpSplit(re *regexp.Regexp, s string, n int) []string {
	if n == 0 {
		return nil
	}

	if len(re.String()) > 0 && len(s) == 0 {
		return []string{""}
	}

	var matches [][]int
	if len(re.SubexpNames()) > 1 {
		matches = re.FindAllStringSubmatchIndex(s, n)
	} else {
		matches = re.FindAllStringIndex(s, n)
	}
	strs := make([]string, 0, len(matches))

	begin, end := 0, 0
	for _, match := range matches {
		if n > 0 && len(strs) >= n-1 {
			break
		}

		end = match[0]
		if match[1] != 0 {
			strs = append(strs, s[begin:end])
		}
		begin = match[1]
		if len(re.SubexpNames()) > 1 {
			strs = append(strs, s[match[0]:match[1]])
		}
	}

	if end != len(s) {
		strs = append(strs, s[begin:])
	}

	return strs
}
