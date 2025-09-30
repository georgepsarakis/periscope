package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/georgepsarakis/periscope/app"
	"github.com/georgepsarakis/periscope/repository"
)

type ProjectHandler struct {
	application app.App
	validate    *validator.Validate
}

func NewProjectHandler(application app.App) ProjectHandler {
	return ProjectHandler{
		application: application,
		validate:    validator.New(validator.WithRequiredStructEnabled()),
	}
}

type ProjectCreateRequest struct {
	Name string `json:"name" validate:"required"`
}

func (h ProjectHandler) Read(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	inputID := chi.URLParam(r, "id")
	id, err := strconv.Atoi(inputID)
	prj, err := h.application.Repository.ProjectFindByID(ctx, uint(id))
	if err != nil {
		err := NewZapError(err, zap.String("project_id", inputID))
		w.Write(NewServerError(ctx, err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	b, err := json.Marshal(
		ProjectReadResponse{
			Project: Project{
				ID:               prj.ID,
				Name:             prj.Name,
				PublicID:         prj.PublicID,
				IngestionAPIKeys: []string{prj.ProjectIngestionAPIKeys[0].Key},
				CreatedAt:        prj.CreatedAt,
				UpdatedAt:        prj.UpdatedAt,
			},
		})
	w.Write(b)
	w.WriteHeader(http.StatusOK)
}

func (h ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req ProjectCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(NewJSONError("json decoding failed", ErrorCodeJSONDecodingFailed))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(NewJSONError("validation failed", ErrorCodeValidationFailed))
		return
	}

	prj, err := h.application.Repository.ProjectCreate(ctx, req.Name)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	b, err := json.Marshal(
		ProjectCreateResponse{
			Project: Project{
				ID:               prj.ID,
				Name:             prj.Name,
				PublicID:         prj.PublicID,
				IngestionAPIKeys: []string{prj.ProjectIngestionAPIKeys[0].Key},
				CreatedAt:        prj.CreatedAt,
				UpdatedAt:        prj.UpdatedAt,
			},
		})
	w.Write(b)
	w.WriteHeader(http.StatusCreated)
}

type ProjectCreateResponse struct {
	Project Project `json:"project"`
}

type ProjectReadResponse struct {
	Project Project `json:"project"`
}

type AlertHandler struct {
	application app.App
	validate    *validator.Validate
}

func NewAlertHandler(application app.App) AlertHandler {
	return AlertHandler{
		application: application,
		validate:    validator.New(validator.WithRequiredStructEnabled()),
	}
}

func (h AlertHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	paramProjectID := chi.URLParam(r, "project_id")
	//paramStatus := chi.URLParam(r, "status")
	//if status == "" {
	//	status = "pending"
	//}
	projectID, err := strconv.Atoi(paramProjectID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	a, err := h.application.Repository.FindAlerts(ctx,
		uint(projectID),
		repository.ListFilters{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := AlertListResponse{
		Alerts: a,
	}
	b, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

type AlertListResponse struct {
	Alerts []repository.Alert `json:"alerts"`
}
