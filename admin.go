package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(200)
	count := cfg.fileserverHits.Load()
	metricsTemplate := `
	<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>
	`
	hits := fmt.Sprintf(metricsTemplate, count)
	_, err := writer.Write([]byte(hits))
	if err != nil {
		fmt.Println(err)
	}
}

func (cfg *apiConfig) handlerReset(writer http.ResponseWriter, request *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(writer, 403, "endpoint allowed only in development environment")
	}
	err := cfg.dbQueries.ResetDB(request.Context())
	if err != nil {
		respondWithError(writer, 500, "error reseting the database")
	}
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
	_, err = writer.Write([]byte("Database reset\n"))
	if err != nil {
		fmt.Println(err)
	}
}
