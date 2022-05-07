package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

var (
	ErrNotExistsInRequestParams = errors.New("invalid parameter")
	errInvalidParamText         = "invalid %s parameter"
)

func HasKeyInReqParams(r *http.Request, key string) bool {
	return InArray[string](
		[]string{key},
		chi.NewRouteContext().URLParams.Keys,
		false,
	)
}
func ReadParam(r *http.Request, key string) (string, bool, error) {
	var value string
	keyExists := !HasKeyInReqParams(r, key)
	if !keyExists {
		return value, false, ErrNotExistsInRequestParams
	}

	value = chi.URLParam(r, key)
	if value == "" {
		return value, keyExists, fmt.Errorf(errInvalidParamText, key)
	}

	return value, keyExists, nil
}

func ReadUUIDParam(r *http.Request, key string) (uuid.UUID, bool, error) {
	var uid uuid.UUID
	value, keyExists, err := ReadParam(r, key)
	if err != nil {
		return uid, keyExists, err
	}

	err = uid.Scan(value)
	if err != nil || uid.String() == "" {
		return uid, keyExists, fmt.Errorf(errInvalidParamText, key)
	}

	return uid, keyExists, nil
}

func ReadNullUUIDParam(r *http.Request, key string) (uuid.NullUUID, bool, error) {
	var uid uuid.NullUUID
	value, keyExists, err := ReadParam(r, key)
	if err != nil {
		return uid, keyExists, err
	}

	err = uid.Scan(value)
	if err != nil || !uid.Valid || uid.UUID.String() == "" {
		return uid, keyExists, fmt.Errorf(errInvalidParamText, key)
	}

	return uid, keyExists, nil
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
func ReadString(r *http.Request, key string, defaultValue string) (string, bool, error) {
	value, exists, err := ReadParam(r, key)

	if err != nil {
		return defaultValue, exists, err
	}

	return value, exists, nil
}

// ReadBool read true or false values from request
func ReadBool(r *http.Request, key string, defaultValue bool) (bool, bool, error) {
	s, exists, err := ReadParam(r, key)

	if err != nil {
		return defaultValue, exists, err
	}

	if s == "true" || s == "t" || s == "y" || s == "1" {
		return true, exists, nil
	} else if s == "false" || s == "f" || s == "n" || s == "0" {
		return false, exists, nil
	}

	return defaultValue, exists, nil
}

// ReadCSV read string by separating by comma
func ReadCSV(r *http.Request, key string, defaultValue []string) ([]string, bool, error) {
	csv, exists, err := ReadParam(r, key)

	if err != nil {
		return defaultValue, exists, err
	}

	return strings.Split(csv, ","), exists, nil
}

// ReadInt read int from request
func ReadInt(r *http.Request, key string, defaultValue int) (int, bool, error) {
	s, exists, err := ReadParam(r, key)
	if err != nil {
		return defaultValue, exists, err
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue, false, fmt.Errorf("must be an integer value")
	}

	return i, true, nil
}

func ReadFloat(r *http.Request, key string, defaultValue float64) (float64, bool, error) {
	s, exists, err := ReadParam(r, key)
	if err != nil {
		return defaultValue, exists, err
	}

	val, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return defaultValue, false, fmt.Errorf("must be a float value")
	}

	return val, true, nil
}
