package handler

import (
	"avito-pr-service/internal/api"
	"avito-pr-service/internal/int_errors"
	"avito-pr-service/internal/model"
	"avito-pr-service/internal/service"
	"encoding/json"
	"net/http"
	"strings"
)

type TeamHandler struct {
	teamService *service.TeamService
}

func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

func (h *TeamHandler) GetTeamGet(w http.ResponseWriter, r *http.Request, params api.GetTeamGetParams) {
	teamName := strings.TrimSpace(params.TeamName)
	if teamName == "" {
		http.Error(w, "team_name must not be empty", http.StatusBadRequest)
		return
	}

	team, err := h.teamService.GetTeam(r.Context(), params.TeamName)
	if err != nil {
		switch err {
		case int_errors.ErrTeamNotFound:
			WriteJSONError(w, http.StatusNotFound, api.NOTFOUND, "team not found")
			return
		default:
			http.Error(w, "internal error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	apiMembers := make([]api.TeamMember, 0, len(team.Members))
	for _, m := range team.Members {
		apiMembers = append(apiMembers, api.TeamMember{
			UserId:   m.ID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	resp := api.Team{
		TeamName: team.TeamName,
		Members:  apiMembers,
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"team": resp})
}

func (h *TeamHandler) PostTeamAdd(w http.ResponseWriter, r *http.Request) {
	var body api.Team

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	teamName := strings.TrimSpace(body.TeamName)
	if teamName == "" {
		http.Error(w, "team_name must not be empty", http.StatusBadRequest)
		return
	}

	members := make([]*model.User, 0, len(body.Members))

	for i, m := range body.Members {
		uid := strings.TrimSpace(m.UserId)
		username := strings.TrimSpace(m.Username)

		if uid == "" || username == "" {
			http.Error(w,
				"user_id and username must not be empty for member index "+string(rune(i)),
				http.StatusBadRequest)
			return
		}

		members = append(members,
			model.NewUser(uid, username, teamName, m.IsActive),
		)
	}

	team, err := h.teamService.CreateTeam(r.Context(), &model.Team{
		TeamName: teamName,
		Members:  members,
	})
	if err != nil {
		switch err {
		case int_errors.ErrUserHasOpenPullRequests:
			WriteJSONError(w, http.StatusBadRequest, api.PREXISTS, "some users have open PRs")
			return
		default:
			http.Error(w, "internal error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	apiMembers := make([]api.TeamMember, 0, len(team.Members))
	for _, m := range team.Members {
		apiMembers = append(apiMembers, api.TeamMember{
			UserId:   m.ID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	resp := api.Team{
		TeamName: team.TeamName,
		Members:  apiMembers,
	}
	WriteJSON(w, http.StatusCreated, map[string]interface{}{"team": resp})
}
