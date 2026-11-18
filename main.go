package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/JakobPavlich/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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
	dbQueries      *database.Queries
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

func respondWithError(writer http.ResponseWriter, code int, msg string) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(code)
	type ErrorJSON struct {
		Error string `json:"error"`
	}
	errJSON := ErrorJSON{Error: msg}
	data, err := json.Marshal(errJSON)
	if err != nil {
		fmt.Println(err)
	}
	_, err = writer.Write(data)
	if err != nil {
		fmt.Println(err)
	}
}

func respondWithJSON(writer http.ResponseWriter, code int, payload interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(code)
	data, err := json.Marshal(payload)
	if err != nil {
		respondWithError(writer, 500, "error marshalling JSON")
	}
	_, err = writer.Write(data)
	if err != nil {
		respondWithError(writer, 500, "error writing data")
	}
}

func checkWords(text string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(text, " ")
	for i, word := range words {
		for _, badWord := range badWords {
			if strings.ToLower(word) == badWord {
				words[i] = "****"
			}
		}
	}
	return strings.Join(words, " ")
}

func handlerValidate(writer http.ResponseWriter, request *http.Request) {
	type chirp struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(request.Body)
	chrp := chirp{}
	err := decoder.Decode(&chrp)
	if err != nil {
		respondWithError(writer, 500, "Error decoding request")
		return
	}
	if len(chrp.Body) > 140 {
		respondWithError(writer, 400, "Chirp is too long")
		return
	}
	checkedText := checkWords(chrp.Body)
	type success struct {
		CleanedBody string `json:"cleaned_body"`
	}
	respondWithJSON(writer, 200, success{CleanedBody: checkedText})
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	apiCfg.dbQueries = database.New(db)
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
