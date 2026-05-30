package handlers

import (
	"encoding/json"
	"net/http"

	"edu-platform/internal/domain"
	"edu-platform/internal/policy"
)

func (h *Handler) Evaluate(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ActivityID int64  `json:"activity_id"`
		StudentID  int64  `json:"student_id"`
		Score      *int   `json:"score"`
		Currency   int64  `json:"currency"`
		Comment    string `json:"comment"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)

	if in.ActivityID == 0 || in.StudentID == 0 {
		http.Error(w, "activity_id and student_id required", http.StatusBadRequest)
		return
	}
	if in.Score == nil {
		http.Error(w, "score is required", http.StatusBadRequest)
		return
	}
	score := *in.Score
	if score < 0 || score > 10 {
		http.Error(w, "score must be between 0 and 10", http.StatusBadRequest)
		return
	}

	evaluatorID := r.Context().Value(ctxUserKey{}).(int64)

	currency := in.Currency
	credits := 0.0
	if currency == 0 {
		currency, credits = policy.ComputeReward(score)
	}

	ev := domain.Evaluation{
		ActivityID:  in.ActivityID,
		EvaluatorID: evaluatorID,
		Score:       score,
		Currency:    currency,
		Credits:     credits,
		Comment:     in.Comment,
	}
	id, err := h.st.CreateEvaluation(r.Context(), ev)
	if err != nil {
		http.Error(w, "create evaluation: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_ = h.st.UpdateActivityStatus(r.Context(), in.ActivityID, "approved")

	if currency != 0 {
		t := domain.Transaction{UserID: in.StudentID, Amount: currency, Reason: "evaluation reward"}
		_, _ = h.st.AddTransaction(r.Context(), t)
	}
	json.NewEncoder(w).Encode(map[string]any{"evaluation_id": id})
}
