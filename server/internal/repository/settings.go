package repository

import (
	"context"
	"database/sql"
)

type Settings struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	OSVEnabled bool    `json:"osvEnabled"`
	OSVURL     *string `json:"osvUrl"`
}

type SettingsUpdate struct {
	Name        *string
	Description *string
	OSVEnabled  *bool
	OSVURL      *string
}

type SettingsRepository struct {
	db *sql.DB
}

const settingsColumns = `name, description, osv_enabled, osv_url`

func scanSettings(s scanner) (Settings, error) {
	var st Settings
	var name, description, osvURL sql.NullString
	if err := s.Scan(&name, &description, &st.OSVEnabled, &osvURL); err != nil {
		return Settings{}, mapErr(err)
	}
	if name.Valid {
		st.Name = &name.String
	}
	if description.Valid {
		st.Description = &description.String
	}
	if osvURL.Valid {
		st.OSVURL = &osvURL.String
	}
	return st, nil
}

func (r *SettingsRepository) Get(ctx context.Context) (Settings, error) {
	return scanSettings(r.db.QueryRowContext(ctx, `SELECT `+settingsColumns+` FROM settings WHERE id = 1`))
}

func (r *SettingsRepository) Update(ctx context.Context, u SettingsUpdate) (Settings, error) {
	return scanSettings(r.db.QueryRowContext(ctx, `
		UPDATE settings SET
			name = CASE WHEN ? THEN ? ELSE name END,
			description = CASE WHEN ? THEN ? ELSE description END,
			osv_enabled = CASE WHEN ? THEN ? ELSE osv_enabled END,
			osv_url = CASE WHEN ? THEN ? ELSE osv_url END
		WHERE id = 1 RETURNING `+settingsColumns,
		u.Name != nil, toNullString(u.Name),
		u.Description != nil, toNullString(u.Description),
		u.OSVEnabled != nil, u.OSVEnabled != nil && *u.OSVEnabled,
		u.OSVURL != nil, toNullString(u.OSVURL)))
}
