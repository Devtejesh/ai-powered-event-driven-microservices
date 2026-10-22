package handler

import (
	"example.com/go-commerce-platform/user-service/models"
	"example.com/go-commerce-platform/user-service/service"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type UserHandler struct {
	Service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{
		Service: service,
	}
}
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	// if r.Method != http.MethodPost {
	// 	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	// 	return
	// }
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request Body", http.StatusBadRequest)
		return
	}
	err = h.Service.CreateUser(r.Context(), &user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetUserById(w http.ResponseWriter, r *http.Request) {
	// if r.Method != http.MethodGet {
	// 	http.Error(w, "Invalid Method", http.StatusMethodNotAllowed)
	// 	return
	// }
	idStr := strings.TrimPrefix(r.URL.Path, "/users/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	user, err := h.Service.GetUserById(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(user)
}
func (h *UserHandler) DeleteUserBYId(w http.ResponseWriter, r *http.Request) {
	// if r.Method != http.MethodDelete {
	// 	http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
	// 	return
	// }
	idStr := strings.TrimPrefix(r.URL.Path, "/users/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user id ", http.StatusBadRequest)
		return
	}
	err = h.Service.DeleteUserBYId(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "user deleted successfully",
	})
}
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	name := query.Get("name")

	page := 1
	limit := 10
	var err error
	if value := query.Get("page"); value != "" {
		page, err = strconv.Atoi(value)
		if err != nil {
			http.Error(w, "Invalid page", http.StatusBadRequest)
			return
		}
	}
	if value := query.Get("limit"); value != "" {
		limit, err = strconv.Atoi(value)
		if err != nil {
			http.Error(w, "Invalid limit", http.StatusBadRequest)
			return
		}
	}

	users, err := h.Service.GetAllUsers(
		r.Context(),
		name,
		page,
		limit,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(users); err != nil {
		return
	}

}
