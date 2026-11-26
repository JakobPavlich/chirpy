package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/JakobPavlich/chirpy/internal/auth"
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
		return
	}
	respondWithJSON(writer, 200, data)
}

func (cfg *apiConfig) handlerPostChirp(writer http.ResponseWriter, request *http.Request) {
	type chirp struct {
		Body string `json:"body"`
	}
	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(writer, 500, "error retrieving token")
		fmt.Println(err)
		return
	}
	userIDFromToken, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(writer, 401, "error validating token")
		return
	}
	decoder := json.NewDecoder(request.Body)
	chrp := chirp{}
	err = decoder.Decode(&chrp)
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
	params.UserID = userIDFromToken
	savedChirp, err := cfg.dbQueries.CreateChirp(request.Context(), params)
	if err != nil {
		respondWithError(writer, 500, err.Error())
		return
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
		return
	}
	respondWithJSON(writer, 201, data)
}

func (cfg *apiConfig) handlerChirps(writer http.ResponseWriter, request *http.Request) {
	chirps, err := cfg.dbQueries.GetChirps(request.Context())
	if err != nil {
		respondWithError(writer, 500, "Error retrieving chirps")
		fmt.Println(err)
		return
	}
	var chirpsSlice []chirpDetails
	for _, chirp := range chirps {
		var chirpJSON chirpDetails
		chirpJSON.ID = chirp.ID
		chirpJSON.Created_at = chirp.CreatedAt
		chirpJSON.Updated_at = chirp.UpdatedAt
		chirpJSON.Body = chirp.Body
		chirpJSON.UserID = chirp.UserID
		chirpsSlice = append(chirpsSlice, chirpJSON)
	}
	data, err := json.Marshal(chirpsSlice)
	if err != nil {
		respondWithError(writer, 500, "Error marshaling chirps slice")
		return
	}
	respondWithJSON(writer, 200, data)
}

func (cfg *apiConfig) handlerChirp(writer http.ResponseWriter, request *http.Request) {
	idString := request.PathValue("chirpID")
	id, err := uuid.Parse(idString)
	if err != nil {
		respondWithError(writer, 404, "error parsing uuid")
		return
	}
	chirp, err := cfg.dbQueries.GetChirpByID(request.Context(), id)
	if err != nil {
		respondWithError(writer, 404, "error retrieving chirp")
		return
	}
	var chirpJSON chirpDetails
	chirpJSON.ID = chirp.ID
	chirpJSON.Created_at = chirp.CreatedAt
	chirpJSON.Updated_at = chirp.UpdatedAt
	chirpJSON.Body = chirp.Body
	chirpJSON.UserID = chirp.UserID
	data, err := json.Marshal(chirpJSON)
	if err != nil {
		respondWithError(writer, 500, "error marshaling chirp")
		return
	}
	respondWithJSON(writer, 200, data)
}

func (cfg *apiConfig) handlerCreateUser(writer http.ResponseWriter, request *http.Request) {
	type newUser struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	decoder := json.NewDecoder(request.Body)
	newUserJSON := newUser{}
	err := decoder.Decode(&newUserJSON)
	if err != nil {
		log.Fatal(err)
	}
	err = request.Body.Close()
	if err != nil {
		fmt.Println(err)
	}
	if newUserJSON.Email == "" {
		respondWithError(writer, 500, "Email required")
		return
	}
	if newUserJSON.Password == "" {
		respondWithError(writer, 500, "Password required")
		return
	}

	var params database.CreateUserParams
	hashed_password, err := auth.HashPassword(newUserJSON.Password)
	if err != nil {
		respondWithError(writer, 500, "Registration failed")
		fmt.Printf("Failed to save password for user '%s': %v", params.Email, err)
	}
	params.Email = newUserJSON.Email
	params.HashedPassword.String = hashed_password
	params.HashedPassword.Valid = true
	user, err := cfg.dbQueries.CreateUser(request.Context(), params)
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
		return
	}

	respondWithJSON(writer, 201, data)
}

func (cfg *apiConfig) handlerLogin(writer http.ResponseWriter, request *http.Request) {
	type user struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	decoder := json.NewDecoder(request.Body)
	userJSON := user{}
	err := decoder.Decode(&userJSON)
	if err != nil {
		respondWithError(writer, 500, "Error checking user.Please try again")
		fmt.Println(err)
		return
	}
	err = request.Body.Close()
	if err != nil {
		fmt.Println(err)
	}
	if userJSON.Email == "" {
		respondWithError(writer, 500, "Email required")
		return
	}
	if userJSON.Password == "" {
		respondWithError(writer, 500, "Password required")
		return
	}

	userDB, err := cfg.dbQueries.GetUserByEmail(request.Context(), userJSON.Email)
	if err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error retrieving user '%s'", userJSON.Email))
		fmt.Println(err)
		return
	}
	ok, err := auth.CheckPasswordHash(userJSON.Password, userDB.HashedPassword.String)
	if err != nil {
		respondWithError(writer, 500, "error checking credentials")
		return
	}
	if !ok {
		respondWithError(writer, 401, "invalid credentials")
		fmt.Println("Invalid credentials")
		return
	}

	var loggedInUser userDetails
	loggedInUser.ID = userDB.ID
	loggedInUser.Email = userDB.Email
	loggedInUser.Created_at = userDB.CreatedAt
	loggedInUser.Updated_at = userDB.UpdatedAt

	expiration := 1 * time.Hour

	token, err := auth.MakeJWT(loggedInUser.ID, cfg.secret, expiration)
	if err != nil {
		respondWithError(writer, 500, "error making token")
		fmt.Println(err)
	}
	loggedInUser.Token = token

	refreshToken := auth.MakeRefreshToken()
	var tokenParams database.CreateRefreshTokenParams
	tokenParams.Token = refreshToken
	tokenParams.ExpiresAt = time.Now().Add(60 * 24 * time.Hour)
	tokenParams.UserID = loggedInUser.ID
	_, err = cfg.dbQueries.CreateRefreshToken(request.Context(), tokenParams)
	if err != nil {
		respondWithError(writer, 500, "error making refresh token")
		fmt.Println(err)
		return
	}
	loggedInUser.RefreshToken = refreshToken

	data, err := json.Marshal(loggedInUser)
	if err != nil {
		respondWithError(writer, 500, "error returning user details")
		fmt.Println(err)
	}
	respondWithJSON(writer, 200, data)
}

func (cfg *apiConfig) handlerRefresh(writer http.ResponseWriter, request *http.Request) {
	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(writer, 500, "error geting refresh token")
		fmt.Println(err)
		return
	}
	dbToken, err := cfg.dbQueries.GetRefreshToken(request.Context(), token)
	if err != nil {
		respondWithError(writer, 401, "error retrieving token from DB")
		fmt.Println(err)
		return
	}
	if expired := time.Now().After(dbToken.ExpiresAt); expired {
		respondWithError(writer, 401, "token expired")
		fmt.Println("token expired")
		return
	}
	if dbToken.RevokedAt.Valid {
		respondWithError(writer, 401, "token revoked")
		fmt.Println("token revoked")
		return
	}
	userID, err := cfg.dbQueries.GetUserFromRefreshToken(request.Context(), token)
	if err != nil {
		respondWithError(writer, 500, "")
		fmt.Println(err)
		return
	}
	jwt, err := auth.MakeJWT(userID, cfg.secret, 1*time.Hour)
	if err != nil {
		respondWithError(writer, 500, "error making JWT")
		fmt.Println(err)
		return
	}
	type tokenReturn struct {
		Token string `json:"token"`
	}
	var returnJWT tokenReturn
	returnJWT.Token = jwt
	data, err := json.Marshal(returnJWT)
	if err != nil {
		respondWithError(writer, 500, "error marshlling jwt")
		fmt.Println(err)
		return
	}
	respondWithJSON(writer, 200, data)
}
