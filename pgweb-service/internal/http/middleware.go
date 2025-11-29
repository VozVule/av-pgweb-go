package httpx

import "net/http"

// WithCORS wraps an http.Handler to add permissive CORS headers for frontend requests.
func WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS,PUT,DELETE")
		w.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type, Accept, HX-Request, HX-Trigger, HX-Target, HX-Current-URL",
		)

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
