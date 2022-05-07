package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func ReadIDParam(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

func ReadUUIDParam(r *http.Request) (uuid.UUID, error) {
	id := chi.URLParam(r, "id")
	var uid uuid.UUID

	if id == "" {
		return uid, errors.New("invalid id parameter")
	}

	err := uid.Scan(id)
	if err != nil || uid.String() == "" {
		return uid, errors.New("invalid id parameter")
	}

	return uid, nil
}

func ReadUUIDParamByKey(r *http.Request, key string) (uuid.UUID, error) {
	var uid uuid.UUID
	if key == "" {
		key = "id"
	}

	id := chi.URLParam(r, key)
	if id == "" {
		return uid, fmt.Errorf("invalid %s parameter", key)
	}

	err := uid.Scan(id)
	if err != nil || uid.String() == "" {
		return uid, fmt.Errorf("invalid %s parameter", key)
	}

	return uid, nil
}

func ReadOptionalUUIDParamByKey(r *http.Request, key string) (uuid.UUID, error) {
	var uid uuid.UUID
	if key == "" {
		key = "id"
	}

	id := chi.URLParam(r, key)
	if id == "" {
		return uid, nil
	}

	err := uid.Scan(id)
	if err != nil || uid.String() == "" {
		return uid, fmt.Errorf("invalid %s parameter", key)
	}

	return uid, nil
}

func ReadNullUUIDParamByKey(r *http.Request, key string) (uuid.NullUUID, error) {
	var uid uuid.NullUUID
	if key == "" {
		key = "id"
	}

	id := chi.URLParam(r, key)
	if id == "" {
		return uid, fmt.Errorf("invalid %s parameter", key)
	}

	err := uid.Scan(id)
	if err != nil || !uid.Valid || uid.UUID.String() == "" {
		return uid, fmt.Errorf("invalid %s parameter", key)
	}

	return uid, nil
}

func ReadJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Use http.MaxBytesReader() to limit the size of the request body to 1MB.
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
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
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// ReadString read string value from request
// return default value if empty string
func ReadString(qs url.Values, key string, defaultValue string) (string, bool) {
	exists := qs.Has(key)
	s := qs.Get(key)

	if s == "" {
		return defaultValue, exists
	}

	return s, exists
}

// ReadBool read true or false values from request
// return default value if value is empty or is not one of true/false values
func ReadBool(qs url.Values, key string, defaultValue bool) (bool, bool) {
	s := qs.Get(key)

	if s == "" {
		return defaultValue, false
	}

	if s == "true" || s == "t" || s == "y" || s == "1" {
		return true, true
	} else if s == "false" || s == "f" || s == "n" || s == "0" {
		return false, true
	}

	return defaultValue, true
}

// ReadCSV read string by separating by comma
// return default value if empty string
func ReadCSV(qs url.Values, key string, defaultValue []string) ([]string, bool) {
	csv := qs.Get(key)

	if csv == "" {
		return defaultValue, false
	}

	return strings.Split(csv, ","), true
}

// ReadInt read int from request
// returns the (value, whether it's valid, error)
// return default value with invalid flag if empty
func ReadInt(qs url.Values, key string, defaultValue int) (int, bool, error) {
	s := qs.Get(key)

	if s == "" {
		return defaultValue, false, nil
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue, false, fmt.Errorf("must be an integer value")
	}

	return i, true, nil
}

func ReadFloat(qs url.Values, key string, defaultValue float64) (float64, bool, error) {
	s := qs.Get(key)

	if s == "" {
		return defaultValue, false, nil
	}

	val, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return defaultValue, false, fmt.Errorf("must be a float value")
	}

	return val, true, nil
}
