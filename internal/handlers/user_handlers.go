package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/example/prreview/internal/models"
	"github.com/example/prreview/internal/repo"
	"github.com/example/prreview/internal/services"
	"github.com/gorilla/mux"
)

func RegisterUserRoutes(r *mux.Router, repos *repo.SQLRepo, svcs *services.Services) {
	r.HandleFunc("/users/setIsActive", func(w http.ResponseWriter, r *http.Request) {
		handleSetIsActive(w, r, repos)
	}).Methods("POST")

	r.HandleFunc("/users/getReview", func(w http.ResponseWriter, r *http.Request) {
		handleGetReview(w, r, repos)
	}).Methods("GET")
}

func handleSetIsActive(w http.ResponseWriter, r *http.Request, repos *repo.SQLRepo) {
	var input struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if input.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	if err := repos.SetUserIsActive(input.UserID, input.IsActive); err != nil {
		http.Error(w, "failed to update user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleGetReview(w http.ResponseWriter, r *http.Request, repos *repo.SQLRepo) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	prs, err := repos.GetPRsForUser(userID)
	if err != nil {
		log.Printf("GetPRsForUser error for user=%q: %v", userID, err)
		http.Error(w, fmt.Sprintf("failed to fetch PRs: %v", err), http.StatusInternalServerError)
		return
	}

	var resp []models.PullRequestShortResp
	for _, pr := range prs {
		resp = append(resp, models.PullRequestShortResp{
			PullRequestID:   pr.PullRequestID,
			PullRequestName: pr.PullRequestName,
			AuthorID:        pr.AuthorID,
			Status:          pr.Status,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string][]models.PullRequestShortResp{"prs": resp}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
