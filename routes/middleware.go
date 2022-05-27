package routes

import "net/http"

func MustHaveSpotifyToken() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			spotifyToken := r.Header.Get("spotify-token")
			if spotifyToken == "" {
				http.Error(w, "missing spotify token", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
