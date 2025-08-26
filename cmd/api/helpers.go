package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"simplewebapi.moviedb/internal/validator"
	"strconv"
	"strings"
)

func (app *application) readIDParam(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id param")
	}
	return id, nil
}

type envelope map[string]interface{}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {

	jsonData, err := json.MarshalIndent(data, "", "  ") //for improved readability only
	if err != nil {
		return err
	}

	jsonData = append(jsonData, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonData)
	return nil
}
func readInt(qs url.Values, key string, defaultVal int, v *validator.Validator) int {
	s := qs.Get(key)
	if s == "" {
		return defaultVal
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultVal
	}
	return i
}
func readString(qs url.Values, key string, defaultVal string) string {
	s := qs.Get(key) //if len qs[key]>1 -> return only qs[0]
	if s == "" {
		return defaultVal
	}
	return s
}
func readCSV(qs url.Values, key string, defaultVal []string) []string {
	csv := qs.Get(key)
	if csv == "" {
		return defaultVal
	}
	return strings.Split(csv, ",")
}
func readListString(qs url.Values, key string, defaultVal []string) []string {
	ls := qs[key]
	if len(ls) == 0 {
		return defaultVal
	}
	var res []string
	copy(res, ls)
	return res
}
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, input interface{}) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	body := json.NewDecoder(r.Body)
	body.DisallowUnknownFields()
	err := body.Decode(input)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}
	err = body.Decode(&struct{}{})
	if err != nil && err != io.EOF {
		return errors.New("invalid input")
	}
	return nil
}
