package job

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateJob(w http.ResponseWriter, r *http.Request) {
	reqp := &CreateJobRequest{}

	// 1. 解析 JSON request body
	if err := json.NewDecoder(r.Body).Decode(reqp); err != nil || strings.TrimSpace(reqp.Kind) == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// 下一步：调用 Service 创建 job
	jobp, err := h.service.createJob(r.Context(), reqp)
	if err != nil {
		http.Error(w, "failed to create job", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Encode should not fail under normal circumstances because Job contains
	// only standard JSON-supported types, and Args comes from a valid JSON request.
	json.NewEncoder(w).Encode(jobp) // TODO：现在是默认不会失败。
}