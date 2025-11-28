package handler

import (
	"net/http"

	"github.com/catalog-service/internal/service"
	"github.com/catalog-service/pkg/response"
)

type AttributesHandler struct {
	service *service.AttributesService
}

func NewAttributesHandler(s *service.AttributesService) *AttributesHandler {
	return &AttributesHandler{service: s}
}

func (h *AttributesHandler) GetAttributes(w http.ResponseWriter, r *http.Request) {
	attributes, err := h.service.GetAttributes()
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	response.JSON(w, http.StatusOK, attributes)
}
