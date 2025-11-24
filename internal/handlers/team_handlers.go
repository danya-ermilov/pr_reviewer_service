package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/example/prreview/internal/models"
	"github.com/example/prreview/internal/repo"
	"github.com/example/prreview/internal/services"
	"github.com/gorilla/mux"
)

func RegisterTeamRoutes(r *mux.Router, repos *repo.SQLRepo, svcs *services.Services) {
	r.HandleFunc("/team/add", func(w http.ResponseWriter, r *http.Request) {
		handleTeamAdd(w, r, repos)
	}).Methods("POST")
	r.HandleFunc("/team/get", func(w http.ResponseWriter, r *http.Request) {
		handleTeamGet(w, r, repos)
	}).Methods("GET")
}

func handleTeamGet(w http.ResponseWriter, r *http.Request, repos *repo.SQLRepo) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		http.Error(w, "team_name required", http.StatusBadRequest)
		return
	}

	team, err := repos.GetTeamByName(teamName)
	if err != nil {
		sendAPIError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	resp := models.TeamResp{
		TeamName: team.TeamName,
		Members:  team.Members,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]models.TeamResp{"team": resp})
}

func handleTeamAdd(w http.ResponseWriter, r *http.Request, repos *repo.SQLRepo) {
	var in struct {
		TeamName string                  `json:"team_name"`
		Members  []models.TeamMemberResp `json:"members"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if in.TeamName == "" {
		sendAPIError(w, http.StatusBadRequest, "TEAM_EXISTS", "team_name is required")
		return
	}

	team, err := repos.CreateTeam(in.TeamName, in.Members)
	if err != nil {
		sendAPIError(w, http.StatusBadRequest, "TEAM_EXISTS", err.Error())
		return
	}

	resp := models.TeamResp{
		TeamName: team.TeamName,
		Members:  team.Members,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]models.TeamResp{"team": resp})
}
