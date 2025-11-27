package util

import (
	"encoding/json"
	"net/http"
)

/*
Write JSON write the given data as a JSON response witht the given status code.
You can write any data that can be marshaled into JSON.
To write JSON reponse use the map[string]any type for arbitrary JSON objects.
*/

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func DecodeJsonBody(request *http.Request) *json.Decoder {
	dec := json.NewDecoder(request.Body)
	dec.DisallowUnknownFields()
	return dec
}
