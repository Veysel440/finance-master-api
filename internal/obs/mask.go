package obs

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

var emailRe = regexp.MustCompile(`(?i)^([^@]{1,3}).*@(.+)$`)

func MaskEmail(s string) string {
	if m := emailRe.FindStringSubmatch(s); m != nil {
		return m[1] + "***@" + m[2]
	}
	return s
}

func HashPII(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:8]) // kısa parmak izi
}

// Bilinen anahtarları maskele
func MaskPIIMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := map[string]any{}
	for k, v := range in {
		lk := strings.ToLower(k)
		switch lk {
		case "email":
			if s, ok := v.(string); ok {
				out[k] = MaskEmail(s)
				break
			}
			out[k] = v
		case "deviceid", "ip":
			if s, ok := v.(string); ok {
				out[k] = HashPII(s)
				break
			}
			out[k] = v
		default:
			out[k] = v
		}
	}
	return out
}
