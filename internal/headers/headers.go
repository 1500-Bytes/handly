package headers

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var (
	ERR_BAD_HEADER = fmt.Errorf("bad header")
	SEPARATOR      = []byte("\r\n")
)

// isTokenValid validates the token(field-name) as defined in RFC 9110
func isTokenValid(str []byte) bool {
	for _, ch := range str {
		found := false
		if ch >= 'A' && ch <= 'Z' ||
			ch >= 'a' && ch <= 'z' ||
			ch >= '0' && ch <= '9' {
			found = true
		}

		switch ch {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			found = true
		}

		if !found {
			return false
		}
	}

	return true
}

type Headers struct {
	headers map[string]string
}

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

func (r *Headers) ForEach(cb func(n, v string)) {
	keys := make([]string, 0, len(r.headers))
	for k := range r.headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		cb(k, r.headers[k])
	}
}

func (h *Headers) GetInt(name string, defaultValue int) int {
	value, exists := h.Get(name)
	if !exists {
		return defaultValue
	}

	v, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return v
}

func (h *Headers) Get(name string) (string, bool) {
	v, ok := h.headers[strings.ToLower(name)]
	if !ok {
		return "", false
	}

	return v, ok
}

func (h *Headers) Set(name, value string) {
	fieldName := strings.ToLower(name)
	if v, ok := h.headers[fieldName]; ok {
		h.headers[fieldName] = fmt.Sprintf("%s,%s", v, value)
	} else {
		h.headers[fieldName] = value
	}
}

func (h *Headers) Replace(name, value string) {
	fieldName := strings.ToLower(name)
	h.headers[fieldName] = value
}

func (h *Headers) Delete(name string) {
	fieldName := strings.ToLower(name)
	delete(h.headers, fieldName)
}

func (h *Headers) parseHeader(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", ERR_BAD_HEADER
	}

	fieldName := parts[0]
	fieldValue := bytes.TrimSpace(parts[1])
	if ok := bytes.HasSuffix(fieldName, []byte(" ")); ok {
		return "", "", ERR_BAD_HEADER
	}

	return string(fieldName), string(fieldValue), nil
}

func (h *Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false

	for {
		idx := bytes.Index(data[read:], SEPARATOR)
		if idx == -1 {
			break
		}

		// Empty header
		if idx == 0 {
			done = true
			read += len(SEPARATOR)
			break
		}

		name, value, err := h.parseHeader(data[read : read+idx])
		if err != nil {
			return 0, false, err
		}

		if !isTokenValid([]byte(name)) {
			return 0, false, ERR_BAD_HEADER
		}

		read += idx + len(SEPARATOR)
		h.Set(name, value)
	}

	return read, done, nil
}
