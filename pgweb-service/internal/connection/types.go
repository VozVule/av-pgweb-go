package connection

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// IntValue accepts JSON numbers or numeric strings and stores them as an int.
type IntValue int

// UnmarshalJSON implements json.Unmarshaler to support string or numeric input.
func (i *IntValue) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" {
		*i = 0
		return nil
	}

	if trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		if s == "" {
			*i = 0
			return nil
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("invalid number %q: %w", s, err)
		}
		*i = IntValue(n)
		return nil
	}

	var n int
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	*i = IntValue(n)
	return nil
}

// Int converts the value to a builtin int.
func (i IntValue) Int() int { return int(i) }
