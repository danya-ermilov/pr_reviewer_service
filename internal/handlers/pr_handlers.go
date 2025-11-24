package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/example/prreview/internal/models"
	"github.com/example/prreview/internal/repo"
	"github.com/example/prreview/internal/services"
	"github.com/gorilla/mux"
)

func RegisterPRRoutes(r *mux.Router, repos *repo.SQLRepo, svcs *services.Services) {
	r.HandleFunc("/pullRequest/create", makeCreatePRHandler(svcs)).Methods("POST")
	r.HandleFunc("/pullRequest/merge", makeMergePRHandler(svcs)).Methods("POST")
	r.HandleFunc("/pullRequest/reassign", makeReassignHandler(svcs)).Methods("POST")
}

func makeCreatePRHandler(svcs *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			PullRequestID   string `json:"pull_request_id"`
			PullRequestName string `json:"pull_request_name"`
			AuthorID        string `json:"author_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if in.PullRequestID == "" || in.PullRequestName == "" || in.AuthorID == "" {
			http.Error(w, "pull_request_id, pull_request_name and author_id required", http.StatusBadRequest)
			return
		}

		pr, err := svcs.PR.CreatePR(in.PullRequestID, in.PullRequestName, in.AuthorID)
		if err != nil {
			if err == services.ErrPRExists {
				sendAPIError(w, http.StatusConflict, "PR_EXISTS", "PR id already exists")
				return
			}
			if err == services.ErrAuthorMissing {
				sendAPIError(w, http.StatusNotFound, "NOT_FOUND", "author not found or has no team")
				return
			}
			sendAPIError(w, http.StatusInternalServerError, "NOT_FOUND", err.Error())
			return
		}

		resp := models.PullRequestResp{
			PullRequestID:     pr["id"].(string),
			PullRequestName:   pr["title"].(string),
			AuthorID:          pr["author"].(string),
			Status:            pr["status"].(string),
			AssignedReviewers: pr["assigned_reviewers"].([]string),
			Team_name:         pr["team_name"].(string),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]models.PullRequestResp{"pr": resp})
	}
}

func makeMergePRHandler(svcs *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			PullRequestID string `json:"pull_request_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if in.PullRequestID == "" {
			http.Error(w, "pull_request_id required", http.StatusBadRequest)
			return
		}
		pr, err := svcs.PR.MergePR(in.PullRequestID)
		if err != nil {
			sendAPIError(w, http.StatusInternalServerError, "NOT_FOUND", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"pr": pr})
	}
}

func makeReassignHandler(svcs *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			PullRequestID string `json:"pull_request_id"`
			OldUserID     string `json:"old_reviewer_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if in.PullRequestID == "" || in.OldUserID == "" {
			http.Error(w, "pull_request_id and old_reviewer_id required", http.StatusBadRequest)
			return
		}
		newID, pr, err := svcs.PR.Reassign(in.PullRequestID, in.OldUserID)
		if err != nil {
			sendAPIError(w, http.StatusConflict, "ERROR", err.Error())
			return
		}
		out := map[string]interface{}{"pr": pr, "replaced_by": newID}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}
