package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/robin-vidal/neetrack/internal/auth"
	"github.com/robin-vidal/neetrack/internal/store"
)

const (
	maxUserNameLength = 64
	minUserPasswordLength = 12
)

type managedUserRequest struct {
	Name     *string `json:"name"`
	Password *string `json:"password"`
	IsAdmin  *bool   `json:"is_admin"`
}

func (s *Server) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.UserFromContext(r.Context())
		if !ok || !user.IsAdmin {
			writeJSON(w, http.StatusForbidden, errorResponse{Error: "forbidden", Message: "Accès réservé aux administrateurs."})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.store.AdminUsers(r.Context())
	if err != nil {
		s.writeUserStoreError(w, r, "listing users", err)
		return
	}
	writeJSON(w, http.StatusOK, struct {
		Users []store.ManagedUser `json:"users"`
	}{Users: users})
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var body managedUserRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == nil || body.Password == nil {
		writeUserValidationError(w, "Le nom et le mot de passe sont obligatoires.")
		return
	}
	name, validationMessage := validateUserName(*body.Name)
	if validationMessage != "" {
		writeUserValidationError(w, validationMessage)
		return
	}
	if validationMessage := validateUserPassword(*body.Password); validationMessage != "" {
		writeUserValidationError(w, validationMessage)
		return
	}
	hash, err := auth.HashPassword(*body.Password)
	if err != nil {
		s.writeUserStoreError(w, r, "hashing user password", err)
		return
	}
	isAdmin := body.IsAdmin != nil && *body.IsAdmin
	user, err := s.store.CreateManagedUser(r.Context(), name, hash, isAdmin)
	if errors.Is(err, store.ErrManagedUserNameTaken) {
		writeJSON(w, http.StatusConflict, errorResponse{Error: "name_taken", Message: "Ce nom d'utilisateur est déjà utilisé."})
		return
	}
	if err != nil {
		s.writeUserStoreError(w, r, "creating managed user", err)
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	id, ok := managedUserID(w, r)
	if !ok {
		return
	}
	var body managedUserRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeUserValidationError(w, "Requête invalide.")
		return
	}
	if body.Name == nil && body.Password == nil && body.IsAdmin == nil {
		writeUserValidationError(w, "Aucune modification à enregistrer.")
		return
	}
	current, _ := auth.UserFromContext(r.Context())
	if current.ID == id && body.IsAdmin != nil && !*body.IsAdmin {
		writeJSON(w, http.StatusConflict, errorResponse{Error: "self_demotion", Message: "Un autre administrateur doit modifier ton rôle."})
		return
	}
	if body.Name != nil {
		name, validationMessage := validateUserName(*body.Name)
		if validationMessage != "" {
			writeUserValidationError(w, validationMessage)
			return
		}
		body.Name = &name
	}
	var passwordHash *string
	if body.Password != nil {
		if validationMessage := validateUserPassword(*body.Password); validationMessage != "" {
			writeUserValidationError(w, validationMessage)
			return
		}
		hash, err := auth.HashPassword(*body.Password)
		if err != nil {
			s.writeUserStoreError(w, r, "hashing user password", err)
			return
		}
		passwordHash = &hash
	}
	user, err := s.store.UpdateManagedUser(r.Context(), id, body.Name, passwordHash, body.IsAdmin)
	if errors.Is(err, store.ErrManagedUserNotFound) {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "user_not_found", Message: "Utilisateur introuvable."})
		return
	}
	if errors.Is(err, store.ErrManagedUserNameTaken) {
		writeJSON(w, http.StatusConflict, errorResponse{Error: "name_taken", Message: "Ce nom d'utilisateur est déjà utilisé."})
		return
	}
	if errors.Is(err, store.ErrLastAdmin) {
		writeJSON(w, http.StatusConflict, errorResponse{Error: "last_admin", Message: "Le dernier administrateur doit rester actif."})
		return
	}
	if err != nil {
		s.writeUserStoreError(w, r, "updating managed user", err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id, ok := managedUserID(w, r)
	if !ok {
		return
	}
	current, _ := auth.UserFromContext(r.Context())
	if current.ID == id {
		writeJSON(w, http.StatusConflict, errorResponse{Error: "self_delete", Message: "Tu ne peux pas supprimer ton propre compte depuis cette page."})
		return
	}
	err := s.store.DeleteManagedUser(r.Context(), id)
	if errors.Is(err, store.ErrManagedUserNotFound) {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "user_not_found", Message: "Utilisateur introuvable."})
		return
	}
	if errors.Is(err, store.ErrLastAdmin) {
		writeJSON(w, http.StatusConflict, errorResponse{Error: "last_admin", Message: "Le dernier administrateur doit rester actif."})
		return
	}
	if err != nil {
		s.writeUserStoreError(w, r, "deleting managed user", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) writeUserStoreError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	s.logger.ErrorContext(r.Context(), operation, "error", err)
	writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal_error", Message: "Erreur interne, réessaie plus tard."})
}

func managedUserID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid_user_id", Message: "Identifiant utilisateur invalide."})
		return 0, false
	}
	return id, true
}

func validateUserName(value string) (string, string) {
	name := strings.TrimSpace(value)
	if name == "" || utf8.RuneCountInString(name) > maxUserNameLength {
	return "", "Le nom doit contenir entre 1 et 64 caractères."
	}
	return name, ""
}

func validateUserPassword(value string) string {
	if utf8.RuneCountInString(value) < minUserPasswordLength {
		return "Le mot de passe doit contenir au moins 12 caractères."
	}
	return ""
}

func writeUserValidationError(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid_user", Message: message})
}
