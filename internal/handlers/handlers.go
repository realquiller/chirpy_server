package handlers

import (
	"context"
	"sort"

	"database/sql"

	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/realquiller/chirpy_server/internal/auth"
	"github.com/realquiller/chirpy_server/internal/database"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
	DbQueries      *database.Queries
	Platform       string
	Secret         string
	PolkaKey       string
}

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Password     string    `json:"password"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Set the Content-Type
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// 2. Set the status code
	w.WriteHeader(http.StatusOK)

	// 3. Write the response body
	w.Write([]byte("OK"))
}

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	count := cfg.FileserverHits.Load()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// format the HTML response
	html := fmt.Sprintf(`<html>
	<body>
	  <h1>Welcome, Chirpy Admin</h1>
	  <p>Chirpy has been visited %d times!</p>
	</body>
  	</html>`, count)

	// write the HTML to the response
	fmt.Fprint(w, html)
}

func (cfg *ApiConfig) ResetHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.Platform != "dev" {
		http.Error(w, "Not allowed in production", http.StatusForbidden)
		return
	}

	err := cfg.DbQueries.DeleteAllUsers(context.Background())
	if err != nil {
		log.Printf("Error deleting all users: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	cfg.FileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	html := fmt.Sprintf("Hits have been set to %d\nAll users deleted", cfg.FileserverHits.Load())
	fmt.Fprint(w, html)
}

func (cfg *ApiConfig) NewUserHandler(w http.ResponseWriter, r *http.Request) {
	type UserRequest struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	var req UserRequest

	// Decode JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	hashed_pw, err := auth.HashPassword(req.Password)

	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Couldn't hash a password", http.StatusInternalServerError)
		return
	}

	user, err := cfg.DbQueries.CreateUser(context.Background(), database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashed_pw,
	})
	if err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}, http.StatusCreated)
}

func (cfg *ApiConfig) GetChirpsHandler(w http.ResponseWriter, r *http.Request) {
	author_id := r.URL.Query().Get("author_id")
	sort_input := r.URL.Query().Get("sort")
	sort_asc := true

	if len(sort_input) > 0 {
		if sort_input == "desc" {
			sort_asc = false
		}
	}

	if len(author_id) > 0 {
		// parse the id into uuid
		parsedID, err := uuid.Parse(author_id)

		//check if the id is valid uuid
		if err != nil {
			respondWithError(w, "Invalid author id", http.StatusBadRequest)
			return
		}
		chirps_author, err := cfg.DbQueries.GetChirpsByAuthor(r.Context(), parsedID)

		if err != nil {
			respondWithError(w, "Error getting chirps from GetChirpsByAuthor function", http.StatusInternalServerError)
			return
		}

		chirps_author_list := []Chirp{}

		for _, chirp := range chirps_author {
			chirps_author_list = append(chirps_author_list, Chirp{
				ID:        chirp.ID,
				CreatedAt: chirp.CreatedAt,
				UpdatedAt: chirp.UpdatedAt,
				Body:      chirp.Body,
				UserID:    chirp.UserID,
			})
		}

		if sort_asc {
			sort.Slice(chirps_author_list, func(i, j int) bool {
				return chirps_author_list[i].CreatedAt.Before(chirps_author_list[j].CreatedAt)
			})
		} else {
			sort.Slice(chirps_author_list, func(i, j int) bool {
				return chirps_author_list[j].CreatedAt.Before(chirps_author_list[i].CreatedAt)
			})
		}

		respondWithJSON(w, chirps_author_list, http.StatusOK)
		return
	}

	chirps, err := cfg.DbQueries.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, "Error getting chirps from GetChirps function", http.StatusInternalServerError)
		return
	}

	chirp_list := []Chirp{}

	for _, chirp := range chirps {
		chirp_list = append(chirp_list, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	if sort_asc {
		sort.Slice(chirp_list, func(i, j int) bool {
			return chirp_list[i].CreatedAt.Before(chirp_list[j].CreatedAt)
		})
	} else {
		sort.Slice(chirp_list, func(i, j int) bool {
			return chirp_list[j].CreatedAt.Before(chirp_list[i].CreatedAt)
		})
	}

	respondWithJSON(w, chirp_list, http.StatusOK)

}

func (cfg *ApiConfig) GetChirpHandler(w http.ResponseWriter, r *http.Request) {
	// get id from endpoint path
	id := r.PathValue("chirpid")

	// parse the id into uuid
	parsedID, err := uuid.Parse(id)

	//check if the id is valid uuid
	if err != nil {
		respondWithError(w, "Invalid chirp ID", http.StatusBadRequest)
		return
	}

	// get the chirp
	chirp, err := cfg.DbQueries.GetChirp(context.Background(), parsedID)

	// check for two errors after getting the chirp
	if err != nil {
		// does the chirp even exist?
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, "Chirp not found", http.StatusNotFound)
			return
		}
		// other error
		respondWithError(w, "Error getting chirp from GetChirp function", http.StatusInternalServerError)
		return

	}
	respondWithJSON(w, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}, http.StatusOK)
}

func (cfg *ApiConfig) ChirpHandler(w http.ResponseWriter, r *http.Request) {
	type ChirpRequest struct {
		Body string `json:"body"`
	}

	// 1. Extract and validate token
	tokenStr, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, "Error extracting token", http.StatusUnauthorized)
		return
	}

	if tokenStr == "" {
		respondWithError(w, "Missing or malformed Authorization header", http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateJWT(tokenStr, cfg.Secret)
	if err != nil {
		respondWithError(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// 2. Parse request body
	var chirpReq ChirpRequest
	if err := json.NewDecoder(r.Body).Decode(&chirpReq); err != nil {
		respondWithError(w, "Failed to parse chirp", http.StatusBadRequest)
		return
	}

	// 3. Create chirp in DB
	chirp, err := cfg.DbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   chirpReq.Body,
		UserID: userID,
	})
	if err != nil {
		respondWithError(w, "Failed to create chirp", http.StatusInternalServerError)
		return
	}

	// 4. Return only the required fields in expected format
	respondWithJSON(w, Chirp{
		ID:     chirp.ID,
		Body:   chirp.Body,
		UserID: chirp.UserID,
	}, http.StatusCreated)
}

func (cfg *ApiConfig) LoginHandler(w http.ResponseWriter, r *http.Request) {
	type LoginRequest struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	// Decode JSON body
	var login LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
		respondWithError(w, "Error decoding JSON in LoginRequest", http.StatusInternalServerError)
		return
	}

	user, err := cfg.DbQueries.GetUser(context.Background(), login.Email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, "User not found", http.StatusNotFound)
			return
		} else {
			respondWithError(w, "Error getting user from GetUser function", http.StatusInternalServerError)
			return
		}
	}

	if err := auth.CheckPasswordHash(user.HashedPassword, login.Password); err != nil {
		respondWithError(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.Secret, time.Duration(3600)*time.Second)
	if err != nil {
		respondWithError(w, "Error creating JWT", http.StatusInternalServerError)
		return
	}

	refresh_token_id, err := auth.MakeRefreshToken()

	if err != nil {
		respondWithError(w, "Error creating refresh token", http.StatusInternalServerError)
		return
	}

	refresh_token, err := cfg.DbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refresh_token_id,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * (24 * 60)),
	})

	if err != nil {
		respondWithError(w, "Error creating refresh token", http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refresh_token.Token,
		IsChirpyRed:  user.IsChirpyRed,
	}, http.StatusOK)

}

func (cfg *ApiConfig) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	type UpdateUserRequest struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	// Decode JSON body
	var update_user UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&update_user); err != nil {
		respondWithError(w, "Error decoding JSON in LoginRequest", http.StatusInternalServerError)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil || token == "" {
		respondWithError(w, "Token is missing", http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.Secret)

	if err != nil {
		respondWithError(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	hashed_pw, err := auth.HashPassword(update_user.Password)

	if err != nil {
		respondWithError(w, "Couldn't hash a password", http.StatusInternalServerError)
		return
	}

	err = cfg.DbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userID,
		HashedPassword: hashed_pw,
		Email:          update_user.Email,
	})

	if err != nil {
		respondWithError(w, "Error updating user", http.StatusInternalServerError)
		return
	}

	updated_user, err := cfg.DbQueries.GetUser(r.Context(), update_user.Email)

	if err != nil {
		respondWithError(w, "Error getting user", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, User{
		ID:          updated_user.ID,
		CreatedAt:   updated_user.CreatedAt,
		UpdatedAt:   updated_user.UpdatedAt,
		Email:       updated_user.Email,
		IsChirpyRed: updated_user.IsChirpyRed,
	}, http.StatusOK)
}

func (cfg *ApiConfig) WebhookUpgradeUserHandler(w http.ResponseWriter, r *http.Request) {
	api_key, err := auth.GetAPIKey(r.Header)

	if err != nil {
		respondWithError(w, "Invalid API format", http.StatusUnauthorized)
		return
	}

	if api_key != cfg.PolkaKey {
		respondWithError(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	type WebhookUpgrade struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	var webhook WebhookUpgrade
	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		respondWithError(w, "Error decoding JSON in WebhookUpgrade", http.StatusInternalServerError)
		return
	}

	if webhook.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	user_id, err := uuid.Parse(webhook.Data.UserID)

	if err != nil {
		respondWithError(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	err = cfg.DbQueries.UpgradeUser(r.Context(), user_id)
	if err != nil {
		respondWithError(w, "User not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

func (cfg *ApiConfig) DeleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("chirpid")
	input_chirp, err := uuid.Parse(id)
	if err != nil {
		respondWithError(w, "Invalid chirp ID", http.StatusBadRequest)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil || token == "" {
		respondWithError(w, "Token is missing", http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.Secret)

	if err != nil {
		respondWithError(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	db_chirp, err := cfg.DbQueries.GetChirp(r.Context(), input_chirp)

	if err != nil {
		respondWithError(w, "Chirp wasn't found", http.StatusNotFound)
		return
	}

	if userID != db_chirp.UserID {
		respondWithError(w, "Forbidden", http.StatusForbidden)
		return
	}

	err = cfg.DbQueries.DeleteChirp(r.Context(), input_chirp)

	if err != nil {
		respondWithError(w, "Error deleting chirp", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *ApiConfig) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := cfg.ValidateRefreshToken(r.Context(), r.Header)
	if err != nil {
		respondWithError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	accToken, err := auth.MakeJWT(refreshToken.UserID, cfg.Secret, time.Hour)
	if err != nil {
		respondWithError(w, "Error creating JWT", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, User{Token: accToken}, http.StatusOK)

}

func (cfg *ApiConfig) RevokeHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := cfg.ValidateRefreshToken(r.Context(), r.Header)
	if err != nil {
		respondWithError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = cfg.DbQueries.RevokeRefreshToken(r.Context(), refreshToken.Token)
	if err != nil {
		respondWithError(w, "Failed to revoke token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204, as per spec
}

func (cfg *ApiConfig) ValidateRefreshToken(ctx context.Context, headers http.Header) (database.RefreshToken, error) {
	token, err := auth.GetBearerToken(headers)
	if err != nil || token == "" {
		return database.RefreshToken{}, fmt.Errorf("unauthorized: invalid or missing bearer token")
	}

	rt, err := cfg.DbQueries.GetRefreshToken(ctx, token)
	if err != nil {
		return database.RefreshToken{}, fmt.Errorf("unauthorized: token not found")
	}

	if time.Now().After(rt.ExpiresAt) {
		return database.RefreshToken{}, fmt.Errorf("unauthorized: token expired")
	}

	if rt.RevokedAt.Valid {
		return database.RefreshToken{}, fmt.Errorf("unauthorized: token revoked")
	}

	return rt, nil
}

func respondWithError(w http.ResponseWriter, msg string, code int) {
	respondWithJSON(w, map[string]string{"error": msg}, code)
}

func respondWithJSON(w http.ResponseWriter, data interface{}, code int) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("JSON marshal error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if _, err := w.Write(jsonBytes); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func filterProfanity(body string) string {
	profanities := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Fields(body)

	for i, word := range words {
		normalized := strings.ToLower(word)
		for _, profanity := range profanities {
			if normalized == profanity {
				words[i] = "****"
			}
		}
	}
	return strings.Join(words, " ")
}

// return func(s *State, cmd Command) error {
// 	user, err := s.Db.GetUser(context.Background(), s.Config.CurrentUserName)
// 	if err != nil {
// 		return fmt.Errorf("error getting user: %w", err)
// 	}
// 	return handler(s, cmd, user)
// }

// func (cfg *ApiConfig) LoginHandler(w http.ResponseWriter, r *http.Request) {
// 	type LoginRequest struct {
// 		Password         string `json:"password"`
// 		Email            string `json:"email"`
// 		ExpiresInSeconds *int32 `json:"expires_in_seconds"`
// 	}

// 	expiration := int32(3600)

// 	// Decode JSON body
// 	var login LoginRequest
// 	if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
// 		respondWithError(w, "Error decoding JSON in LoginRequest", http.StatusInternalServerError)
// 		return
// 	}

// 	if login.ExpiresInSeconds != nil {
// 		expires := *login.ExpiresInSeconds
// 		if expires > 0 && expires < 3600 {
// 			expiration = expires
// 		}
// 	}

// 	user, err := cfg.DbQueries.GetUser(context.Background(), login.Email)

// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			respondWithError(w, "User not found", http.StatusNotFound)
// 			return
// 		} else {
// 			respondWithError(w, "Error getting user from GetUser function", http.StatusInternalServerError)
// 			return
// 		}
// 	}

// 	if err := auth.CheckPasswordHash(user.HashedPassword, login.Password); err != nil {
// 		respondWithError(w, "Invalid email or password", http.StatusUnauthorized)
// 		return
// 	}

// 	token, err := auth.MakeJWT(user.ID, cfg.Secret, time.Duration(expiration)*time.Second)
// 	if err != nil {
// 		respondWithError(w, "Error creating JWT", http.StatusInternalServerError)
// 		return
// 	}

// 	refresh_token, err := auth.MakeRefreshToken()

// 	if err != nil {
// 		respondWithError(w, "Error creating refresh token", http.StatusInternalServerError)
// 		return
// 	}
// 	respondWithJSON(w, User{
// 		ID:           user.ID,
// 		CreatedAt:    user.CreatedAt,
// 		UpdatedAt:    user.UpdatedAt,
// 		Email:        user.Email,
// 		Token:        token,
// 		RefreshToken: refresh_token,
// 	}, http.StatusOK)

// }
