package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

func handlerReadiness(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
	_, err := writer.Write([]byte("OK\n"))
	if err != nil {
		fmt.Println(err)
	}
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

var apiCfg apiConfig

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
	cfg.fileserverHits.Store(0)
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
	_, err := writer.Write([]byte("Metrics have been reset\n"))
	if err != nil {
		fmt.Println(err)
	}
}

// TODO: še za narest, za zdaj le osnutek
func handlerValidate(writer http.ResponseWriter, request *http.Request) {
	type chirp struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(request.Body)
	chrp := chirp{}
	err := decoder.Decode(&chrp)
	if err != nil {
		log.Printf("Error decoding chirp: %s", err)
		writer.WriteHeader(500)
		return
	}
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidate)
	var server http.Server
	server.Handler = mux
	server.Addr = ":8080"
	log.Fatal(server.ListenAndServe())
}
