package handlers

import (
	"encoding/json"
	"net/http"

	"edu-platform/internal/domain"
	"edu-platform/internal/policy"
)

type evaluationRequest struct {
	Points  int      `json:"points"`
	Credits *float64 `json:"credits"` // optional; defaults from points
	Comment string   `json:"comment"`
	Reject  bool     `json:"reject"`
}

// Evaluate godoc
// @Summary  Evaluate or reject an activity
// @Description Awards points/credits (status EVALUATED) or rejects with a comment (status REJECTED). group_admin is restricted to their own group.
// @Tags     evaluation
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    id   path int               true "Activity id"
// @Param    body body evaluationRequest true "Evaluation"
// @Success  201 {object} domain.Evaluation
// @Failure  400 {object} errorResponse
// @Failure  403 {object} errorResponse
// @Failure  404 {object} errorResponse
// @Router   /activities/{id}/evaluation [post]
func (h *Handler) Evaluate(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in evaluationRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}

	a, err := h.st.GetActivity(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "activity not found")
		return
	}
	// group_admin may only evaluate activities of their own group.
	if role(r) == "group_admin" && a.StudentGroup != group(r) {
		writeErr(w, http.StatusForbidden, "activity belongs to another group")
		return
	}

	status := domain.StatusEvaluated
	ev := domain.Evaluation{ActivityID: id, AdminID: sub(r), Comment: in.Comment}

	if in.Reject {
		status = domain.StatusRejected
	} else {
		if in.Points < 0 || in.Points > policy.MaxPoints {
			writeErr(w, http.StatusBadRequest, "points must be between 0 and 10")
			return
		}
		ev.Points = in.Points
		if in.Credits != nil {
			ev.Credits = *in.Credits
		} else {
			ev.Credits = policy.CreditsForPoints(in.Points)
		}
	}

	evID, err := h.st.UpsertEvaluation(r.Context(), ev, status)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "save evaluation: "+err.Error())
		return
	}
	ev.ID = evID
	writeJSON(w, http.StatusCreated, ev)
}
