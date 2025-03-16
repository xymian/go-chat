package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func ParseBody(r *http.Request, o interface{}) {
	if body, err := io.ReadAll(r.Body); err == nil {
		if err = json.Unmarshal(body, o); err != nil {
			return
		}
	}
}

func GenerateRoomId(username0 string, username1 string) (string, error) {
	if len(username0) > 0 && len(username1) > 0 {
		if username0[0] <= username1[1] {
			return fmt.Sprint(username0, "+", username1), nil
		} else {
			return fmt.Sprint(username1, "+", username0), nil
		}
	}
	return "", errors.New("invalid username")
}
