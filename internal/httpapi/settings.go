package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/robin-vidal/neetrack/internal/auth"
	"github.com/robin-vidal/neetrack/internal/store"
)

const (
	minTargetRetention = 0.70
	maxTargetRetention = 0.99
	minBlockSize       = 1
	maxBlockSize       = 6
	minOpenMaxSetting  = 1
	maxOpenMaxSetting  = 3
)

type settingsResponse struct {
	TargetRetention float64 `json:"target_retention"`
	TargetDate      *string `json:"target_date"`
	BlockSize       int     `json:"block_size"`
	OpenMax         int     `json:"open_max"`
	NewPct          int     `json:"new_pct"`
	SessionSize     int     `json:"session_size"`
	IncludeHard     bool    `json:"include_hard"`
}

func toSettingsResponse(s store.UserSettings) settingsResponse {
	resp := settingsResponse{
		TargetRetention: s.TargetRetention,
		BlockSize:       s.BlockSize,
		OpenMax:         s.OpenMax,
		NewPct:          s.NewPct,
		SessionSize:     s.SessionSize,
		IncludeHard:     s.IncludeHard,
	}
	if s.TargetDate != nil {
		d := s.TargetDate.Format(dateLayout)
		resp.TargetDate = &d
	}
	return resp
}

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())

	settings, err := s.store.Settings(r.Context(), user.ID)
	if err != nil {
		s.logger.ErrorContext(r.Context(), "loading settings", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error:   "internal_error",
			Message: "Erreur interne, réessaie plus tard.",
		})
		return
	}

	writeJSON(w, http.StatusOK, toSettingsResponse(settings))
}

// patchSettingsRequest patches only the fields present in the request. A
// nil field leaves the current value untouched. TargetDate is a string, not
// *string: an empty string clears the target date, since a JSON null cannot
// be distinguished from an absent key with the standard decoder.
type patchSettingsRequest struct {
	TargetRetention *float64 `json:"target_retention"`
	TargetDate      *string  `json:"target_date"`
	BlockSize       *int     `json:"block_size"`
	OpenMax         *int     `json:"open_max"`
	NewPct          *int     `json:"new_pct"`
	SessionSize     *int     `json:"session_size"`
	IncludeHard     *bool    `json:"include_hard"`
}

func (s *Server) handlePatchSettings(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	ctx := r.Context()

	var body patchSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error:   "invalid_body",
			Message: "Requête invalide.",
		})
		return
	}

	settings, err := s.store.Settings(ctx, user.ID)
	if err != nil {
		s.logger.ErrorContext(ctx, "loading settings", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error:   "internal_error",
			Message: "Erreur interne, réessaie plus tard.",
		})
		return
	}

	if invalidField := applySettingsPatch(&settings, body); invalidField != "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error:   "invalid_settings",
			Message: "Réglage invalide : " + invalidField + ".",
		})
		return
	}

	if err := s.store.UpdateSettings(ctx, user.ID, settings); err != nil {
		s.logger.ErrorContext(ctx, "updating settings", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error:   "internal_error",
			Message: "Erreur interne, réessaie plus tard.",
		})
		return
	}

	writeJSON(w, http.StatusOK, toSettingsResponse(settings))
}

// applySettingsPatch merges body into settings, returning the name of the
// first field that fails validation, or "" if the patch is valid.
func applySettingsPatch(settings *store.UserSettings, body patchSettingsRequest) string {
	if body.TargetRetention != nil {
		if *body.TargetRetention < minTargetRetention || *body.TargetRetention > maxTargetRetention {
			return "target_retention"
		}
		settings.TargetRetention = *body.TargetRetention
	}

	if body.TargetDate != nil {
		if *body.TargetDate == "" {
			settings.TargetDate = nil
		} else {
			parsed, err := time.Parse(dateLayout, *body.TargetDate)
			if err != nil {
				return "target_date"
			}
			settings.TargetDate = &parsed
		}
	}

	if body.BlockSize != nil {
		if *body.BlockSize < minBlockSize || *body.BlockSize > maxBlockSize {
			return "block_size"
		}
		settings.BlockSize = *body.BlockSize
	}

	if body.OpenMax != nil {
		if *body.OpenMax < minOpenMaxSetting || *body.OpenMax > maxOpenMaxSetting {
			return "open_max"
		}
		settings.OpenMax = *body.OpenMax
	}

	if body.NewPct != nil {
		if *body.NewPct < 0 || *body.NewPct > 100 {
			return "new_pct"
		}
		settings.NewPct = *body.NewPct
	}

	if body.SessionSize != nil {
		if *body.SessionSize < minSessionSize || *body.SessionSize > maxSessionSize {
			return "session_size"
		}
		settings.SessionSize = *body.SessionSize
	}

	if body.IncludeHard != nil {
		settings.IncludeHard = *body.IncludeHard
	}

	return ""
}
