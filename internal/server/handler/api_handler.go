package handler

import (
	"avito-pr-service/internal/api"
	"encoding/json"
	"net/http"
)

type APIHandler struct {
	pr   *PRHandler
	user *UserHandler
	team *TeamHandler
	stat *StatisticsHandler
}

func NewAPIHandler(pr *PRHandler, user *UserHandler, team *TeamHandler, stat *StatisticsHandler) *APIHandler {
	return &APIHandler{
		pr:   pr,
		user: user,
		team: team,
		stat: stat,
	}
}

func (h *APIHandler) PostPullRequestCreate(w http.ResponseWriter, r *http.Request) {
	h.pr.PostPullRequestCreate(w, r)
}

func (h *APIHandler) GetTeamGet(w http.ResponseWriter, r *http.Request, params api.GetTeamGetParams) {
	h.team.GetTeamGet(w, r, params)
}

func (h *APIHandler) GetUsersGetReview(w http.ResponseWriter, r *http.Request, params api.GetUsersGetReviewParams) {
	h.user.GetUsersGetReview(w, r, params)
}

func (h *APIHandler) PostPullRequestMerge(w http.ResponseWriter, r *http.Request) {
	h.pr.PostPullRequestMerge(w, r)
}

func (h *APIHandler) PostPullRequestReassign(w http.ResponseWriter, r *http.Request) {
	h.pr.PostPullRequestReassign(w, r)
}

func (h *APIHandler) PostTeamAdd(w http.ResponseWriter, r *http.Request) {
	h.team.PostTeamAdd(w, r)
}

func (h *APIHandler) PostUsersSetIsActive(w http.ResponseWriter, r *http.Request) {
	h.user.PostUsersSetIsActive(w, r)
}

func (h *APIHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	h.stat.GetStatistics(w, r)
}

func WriteJSONError(w http.ResponseWriter, status int, code api.ErrorResponseErrorCode, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := api.ErrorResponse{}
	resp.Error.Code = code
	resp.Error.Message = msg

	_ = json.NewEncoder(w).Encode(resp)
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}
