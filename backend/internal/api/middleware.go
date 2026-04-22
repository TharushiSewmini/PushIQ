package api

import (
    "net/http"

    "github.com/gorilla/mux"
    "github.com/sirupsen/logrus"
)

func apiKeyMiddleware(apiKey string, logger *logrus.Logger) mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if apiKey == "" || r.Header.Get("X-Api-Key") != apiKey {
                logger.WithFields(logrus.Fields{
                    "path":   r.URL.Path,
                    "method": r.Method,
                }).Warn("unauthorized request")
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
