package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/robin-vidal/neetrack/internal/store/storedb"
)

// UserSettings holds one user's session and retention preferences.
type UserSettings struct {
	TargetRetention float64
	TargetDate      *time.Time
	BlockSize       int
	OpenMax         int
	NewPct          int
	SessionSize     int
	IncludeHard     bool
}

// Settings returns userID's settings.
func (s *Store) Settings(ctx context.Context, userID int64) (UserSettings, error) {
	row, err := s.queries.SettingsByUser(ctx, userID)
	if err != nil {
		return UserSettings{}, fmt.Errorf("querying settings: %w", err)
	}

	settings := UserSettings{
		TargetRetention: row.TargetRetention,
		BlockSize:       int(row.BlockSize),
		OpenMax:         int(row.OpenMax),
		NewPct:          int(row.NewPct),
		SessionSize:     int(row.SessionSize),
		IncludeHard:     row.IncludeHard,
	}
	if row.TargetDate.Valid {
		t := row.TargetDate.Time
		settings.TargetDate = &t
	}

	return settings, nil
}

// UpdateSettings replaces userID's settings.
func (s *Store) UpdateSettings(ctx context.Context, userID int64, settings UserSettings) error {
	var targetDate pgtype.Date
	if settings.TargetDate != nil {
		targetDate = pgtype.Date{Time: *settings.TargetDate, Valid: true}
	}

	err := s.queries.UpdateSettings(ctx, storedb.UpdateSettingsParams{
		UserID:          userID,
		TargetRetention: settings.TargetRetention,
		TargetDate:      targetDate,
		BlockSize:       int16(settings.BlockSize),
		OpenMax:         int16(settings.OpenMax),
		NewPct:          int16(settings.NewPct),
		SessionSize:     int16(settings.SessionSize),
		IncludeHard:     settings.IncludeHard,
	})
	if err != nil {
		return fmt.Errorf("updating settings: %w", err)
	}

	return nil
}
