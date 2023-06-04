package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

func LogRequestBodyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		r.Body.Close()
		r.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

		// Attempt to Compact JSON for logging
		tmpBody := &bytes.Buffer{}
		if err := json.Compact(tmpBody, bodyBytes); err == nil {
			bodyBytes = tmpBody.Bytes()
		}
		log.Debug().Str("body", strings.Replace(string(bodyBytes), "\n", "", -1)).Msg("Body Received.")
		next.ServeHTTP(w, r)

	})
}
