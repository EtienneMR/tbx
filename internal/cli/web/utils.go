package web

import (
	"bufio"
	"io"
	"regexp"

	"github.com/EtienneMR/tbx/internal/tui"
)

func extractFirstRegex(r io.Reader, re *regexp.Regexp) ([][]byte, error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		bytes := scanner.Bytes()
		tui.Debug("%s", string(bytes))
		idx := re.FindSubmatchIndex(bytes)
		if idx != nil {
			out := make([][]byte, len(idx)/2)
			for i := 0; i < len(idx); i += 2 {
				part := make([]byte, idx[i+1]-idx[i])
				copy(part, bytes[idx[i]:idx[i+1]])
				out[i/2] = part
			}
			return out, nil
		}
	}

	return nil, nil
}
