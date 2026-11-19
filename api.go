package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/JakobPavlich/chirpy/internal/database"
	"github.com/google/uuid"
)

func handlerReadiness(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
	_, err := writer.Write([]byte("OK\n"))
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

func respondWithJSON(writer http.ResponseWriter, code int, payload []byte) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(code)
	_, err := writer.Write(payload)
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
	request.Body.Close()
	if len(chrp.Body) > 140 {
		respondWithError(writer, 400, "Chirp is too long")
		return
	}
	checkedText := checkWords(chrp.Body)
	type success struct {
		CleanedBody string `json:"cleaned_body"`
	}
	var cleanBody success
	cleanBody.CleanedBody = checkedText
	data, err := json.Marshal(cleanBody)
	if err != nil {
		respondWithError(writer, 500, "error marshalling JSON")
	}
	respondWithJSON(writer, 200, data)
}

func (cfg *apiConfig) handlerChirps(writer http.ResponseWriter, request *http.Request) {
	type chirp struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	decoder := json.NewDecoder(request.Body)
	chrp := chirp{}
	err := decoder.Decode(&chrp)
	if err != nil {
		respondWithError(writer, 500, "Error decoding request")
		return
	}
	request.Body.Close()
	if len(chrp.Body) > 140 {
		respondWithError(writer, 400, "Chirp is too long")
		return
	}
	checkedText := checkWords(chrp.Body)
	var params database.CreateChirpParams
	params.Body = checkedText
	params.UserID = chrp.UserID
	savedChirp, err := cfg.dbQueries.CreateChirp(request.Context(), params)
	if err != nil {
		respondWithError(writer, 500, err.Error())
	}
	var chirpJSON chirpDetails
	chirpJSON.ID = savedChirp.ID
	chirpJSON.Created_at = savedChirp.CreatedAt
	chirpJSON.Updated_at = savedChirp.UpdatedAt
	chirpJSON.Body = savedChirp.Body
	chirpJSON.UserID = savedChirp.UserID
	data, err := json.Marshal(chirpJSON)
	if err != nil {
		respondWithError(writer, 500, "Error marshaling chirp")
	}
	respondWithJSON(writer, 201, data)
}

func (cfg *apiConfig) handlerCreateUser(writer http.ResponseWriter, request *http.Request) {
	type userEmail struct {
		Email string `json:"email"`
	}
	decoder := json.NewDecoder(request.Body)
	uEmail := userEmail{}
	err := decoder.Decode(&uEmail)
	if err != nil {
		fmt.Println(err)
	}
	user, err := cfg.dbQueries.CreateUser(request.Context(), uEmail.Email)
	if err != nil {
		fmt.Println(err)
	}
	var userJSON userDetails
	userJSON.ID = user.ID
	userJSON.Created_at = user.CreatedAt
	userJSON.Updated_at = user.UpdatedAt
	userJSON.Email = user.Email

	data, err := json.Marshal(userJSON)
	if err != nil {
		respondWithError(writer, 500, "error marshalling JSON in handlerCreateUser")
	}

	respondWithJSON(writer, 201, data)
}
