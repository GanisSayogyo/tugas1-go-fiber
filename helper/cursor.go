package helper

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// EncodeCursor mengodekan pasangan created_at dan id.
// Base64 hanya encoding, bukan enkripsi.
func EncodeCursor(createdAt time.Time, id int) string {
	value := fmt.Sprintf("%s|%d", createdAt.UTC().Format(time.RFC3339Nano), id)
	return base64.RawURLEncoding.EncodeToString([]byte(value))
}

// DecodeCursor memvalidasi dan membaca cursor.
func DecodeCursor(cursor string) (time.Time, int, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, 0, fmt.Errorf("cursor bukan base64 yang valid")
	}

	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 {
		return time.Time{}, 0, fmt.Errorf("format cursor tidak valid")
	}

	createdAt, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, 0, fmt.Errorf("waktu pada cursor tidak valid")
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil || id < 1 {
		return time.Time{}, 0, fmt.Errorf("id pada cursor tidak valid")
	}

	return createdAt.UTC(), id, nil
}
