package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)


func readJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxBytes := 1048576

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(data)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = decoder.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only have a single JSON value")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) error {
	out, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(out)

	return nil
}

