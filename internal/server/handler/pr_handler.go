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

type PostPullRequestReassignResponse struct {
	Pr         api.PullRequest `json:"pr"`
	ReplacedBy string          `json:"replaced_by"`
}

type PRHandler struct {
	prService *service.PRService
}

func NewPRHandler(prService *service.PRService) *PRHandler {
	return &PRHandler{prService: prService}
}

func (h *PRHandler) PostPullRequestCreate(w http.ResponseWriter, r *http.Request) {
	var body api.PostPullRequestCreateJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req := model.CreatePRRequest{
		PRID:     strings.TrimSpace(body.PullRequestId),
		PRName:   strings.TrimSpace(body.PullRequestName),
		AuthorID: strings.TrimSpace(body.AuthorId),
	}

	if req.PRID == "" || req.PRName == "" || req.AuthorID == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	pr, err := h.prService.CreatePR(r.Context(), req)
	if err != nil {
		switch err {
		case int_errors.ErrPRExists:
			WriteJSONError(w, http.StatusConflict, api.PREXISTS, "pull request already exists")
			return
		case int_errors.ErrUserNotFound:
			WriteJSONError(w, http.StatusNotFound, api.NOTFOUND, "author not found")
			return
		default:
			http.Error(w, "internal error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	status := api.PullRequestStatusOPEN
	if pr.Status == model.StatusMerged {
		status = api.PullRequestStatusMERGED
	}

	resp := api.PullRequest{
		PullRequestId:     pr.PRID,
		PullRequestName:   pr.PRName,
		AuthorId:          pr.AuthorID,
		AssignedReviewers: pr.AssignedReviewers,
		Status:            status,
		CreatedAt:         &pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
	WriteJSON(w, http.StatusCreated, map[string]interface{}{"pr": resp})
}

func (h *PRHandler) PostPullRequestMerge(w http.ResponseWriter, r *http.Request) {
	var body api.PostPullRequestMergeJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	prID := strings.TrimSpace(body.PullRequestId)
	if prID == "" {
		http.Error(w, "pull_request_id must not be empty", http.StatusBadRequest)
		return
	}

	pr, err := h.prService.MergePR(r.Context(), prID)
	if err != nil {
		switch err {
		case int_errors.ErrPRNotFound:
			WriteJSONError(w, http.StatusNotFound, api.NOTFOUND, "pull request not found")
			return
		default:
			http.Error(w, "internal error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	status := api.PullRequestStatusOPEN
	if pr.Status == model.StatusMerged {
		status = api.PullRequestStatusMERGED
	}

	resp := api.PullRequest{
		PullRequestId:     pr.PRID,
		PullRequestName:   pr.PRName,
		AuthorId:          pr.AuthorID,
		AssignedReviewers: pr.AssignedReviewers,
		Status:            status,
		CreatedAt:         &pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"pr": resp})
}

func (h *PRHandler) PostPullRequestReassign(w http.ResponseWriter, r *http.Request) {
	var body api.PostPullRequestReassignJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	prID := strings.TrimSpace(body.PullRequestId)
	oldReviewerID := strings.TrimSpace(body.OldUserId)
	if prID == "" || oldReviewerID == "" {
		http.Error(w, "pull_request_id and old_user_id must not be empty", http.StatusBadRequest)
		return
	}

	pr, newReviewerID, err := h.prService.ReassignReviewer(r.Context(), prID, oldReviewerID)
	if err != nil {
		switch err {
		case int_errors.ErrPRNotFound:
			WriteJSONError(w, http.StatusNotFound, api.NOTFOUND, "pull request not found")
			return
		case int_errors.ErrReviewerNotAssigned:
			WriteJSONError(w, http.StatusConflict, api.NOTASSIGNED, "reviewer not assigned to PR")
			return
		case int_errors.ErrPRMerged:
			WriteJSONError(w, http.StatusConflict, api.PRMERGED, "cannot reassign merged PR")
			return
		case int_errors.ErrNoReplacementCandidate:
			WriteJSONError(w, http.StatusConflict, api.NOCANDIDATE, "no replacement candidate available")
			return
		case int_errors.ErrUserNotFound:
			WriteJSONError(w, http.StatusNotFound, api.NOTFOUND, "user not found")
			return
		}
	}

	status := api.PullRequestStatusOPEN
	if pr.Status == model.StatusMerged {
		status = api.PullRequestStatusMERGED
	}
	resp := PostPullRequestReassignResponse{
		Pr: api.PullRequest{
			PullRequestId:     pr.PRID,
			PullRequestName:   pr.PRName,
			AuthorId:          pr.AuthorID,
			AssignedReviewers: pr.AssignedReviewers,
			Status:            status,
			CreatedAt:         &pr.CreatedAt,
			MergedAt:          pr.MergedAt,
		},
		ReplacedBy: newReviewerID,
	}

	WriteJSON(w, http.StatusOK, resp)
}
