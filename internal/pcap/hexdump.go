package pcap

import (
	"fmt"
	"strings"
)

func HexToCPS(hexStr string) string {
	cleaned := strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			return r
		}
		return -1
	}, hexStr)

	if len(cleaned) == 0 {
		return ""
	}

	if len(cleaned) < 64 {
		return ""
	}

	return fmt.Sprintf("<b 0x%s>", cleaned)
}
