package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/JakobPavlich/chirpy/internal/auth"
	"github.com/JakobPavlich/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUpdateUserDetails(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(401)
		return
	}
	type details struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	var uDetails details
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&uDetails)
	if err != nil {
		fmt.Println(err)
		respondWithError(w, 500, "")
		return
	}
	id, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(401)
		return
	}
	hashedPassword, err := auth.HashPassword(uDetails.Password)
	if err != nil {
		fmt.Println(err)
		respondWithError(w, 500, "")
		return
	}
	var params database.UpdateUserEmailAndPasswordParams
	params.ID = id
	params.Email = uDetails.Email
	params.HashedPassword.String = hashedPassword
	params.HashedPassword.Valid = true
	params.UpdatedAt = time.Now().UTC()
	user, err := cfg.dbQueries.UpdateUserEmailAndPassword(r.Context(), params)
	if err != nil {
		fmt.Println(err)
		respondWithError(w, 500, "")
		return
	}
	type response struct {
		ID          uuid.UUID `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}
	var res response
	res.ID = user.ID
	res.CreatedAt = user.CreatedAt
	res.UpdatedAt = user.UpdatedAt
	res.Email = user.Email
	res.IsChirpyRed = user.IsChirpyRed.Bool
	resJSON, err := json.Marshal(res)
	if err != nil {
		fmt.Println(err)
		respondWithError(w, 500, "")
		return
	}

	respondWithJSON(w, 200, resJSON)
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(401)
		return
	}
	userIDFromToken, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(401)
		return
	}
	chirpIDString := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	chirp, err := cfg.dbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	if chirp.UserID != userIDFromToken {
		w.WriteHeader(403)
		return
	}
	err = cfg.dbQueries.DeleteChirp(r.Context(), chirp.ID)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(204)
}

func (cfg *apiConfig) handlerUserUpgrade(w http.ResponseWriter, r *http.Request) {
	APIKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		w.WriteHeader(401)
		fmt.Println(err)
		return
	}
	if APIKey != cfg.polkaKey {
		w.WriteHeader(401)
		fmt.Println("wrong API key")
		return
	}
	type upgradeData struct {
		UserID string `json:"user_id"`
	}
	type upgradeRequest struct {
		Event string      `json:"event"`
		Data  upgradeData `json:"data"`
	}
	var userRequest upgradeRequest
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&userRequest)
	if err != nil {
		w.WriteHeader(500)
		fmt.Println(err)
		return
	}
	if userRequest.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}
	id, err := uuid.Parse(userRequest.Data.UserID)
	if err != nil {
		w.WriteHeader(404)
		fmt.Println(err)
		return
	}
	err = cfg.dbQueries.ChirpyRedUpgrade(r.Context(), id)
	if err != nil {
		w.WriteHeader(404)
		fmt.Println(err)
		return
	}
	w.WriteHeader(204)
}
