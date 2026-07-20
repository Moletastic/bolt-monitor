package notifications

import "encoding/base64"

func base64Encode(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}
