package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year"`
	Runtime   Runtime   `json:"runtime,omitempty"`
	Genres    []string  `json:"genres"`
	Version   int32     `json:"version"`
}

var ErrInvalidRuntimeFormat = errors.New("invalid Runtime field format") //Runtime tên riêng

type Runtime int32

func (r Runtime) MarshalJSON() ([]byte, error) {
	rString := fmt.Sprintf("%d mins", r)

	quotedValue := strconv.Quote(rString)

	return []byte(quotedValue), nil
}
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))

	if err != nil {
		return ErrInvalidRuntimeFormat
	}
	parts := strings.Split(unquotedJSONValue, " ")
	if len(parts) != 2 || (parts[1] != "mins" && parts[1] != "minutes") {
		return ErrInvalidRuntimeFormat
	}
	val, err := strconv.ParseInt(parts[0], 10, 32)

	if err != nil {
		return ErrInvalidRuntimeFormat
	}
	*r = Runtime(val)

	return nil
}
