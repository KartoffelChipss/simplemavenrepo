package repository

import (
	"context"
	"database/sql"
)

type Settings struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type SettingsRepository struct {
	db *sql.DB
}

const settingsColumns = `name, description`

func scanSettings(s scanner) (Settings, error) {
	var st Settings
	var name, description sql.NullString
	if err := s.Scan(&name, &description); err != nil {
		return Settings{}, mapErr(err)
	}
	if name.Valid {
		st.Name = &name.String
	}
	if description.Valid {
		st.Description = &description.String
	}
	return st, nil
}

func (r *SettingsRepository) Get(ctx context.Context) (Settings, error) {
	return scanSettings(r.db.QueryRowContext(ctx, `SELECT `+settingsColumns+` FROM settings WHERE id = 1`))
}

func (r *SettingsRepository) Update(ctx context.Context, name, description *string) (Settings, error) {
	return scanSettings(r.db.QueryRowContext(ctx, `
		UPDATE settings SET
			name = CASE WHEN ? THEN ? ELSE name END,
			description = CASE WHEN ? THEN ? ELSE description END
		WHERE id = 1 RETURNING `+settingsColumns,
		name != nil, toNullString(name), description != nil, toNullString(description)))
}
