package audit

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository interface {
	Record(ctx context.Context, dto CreateEventDTO) error
	GetAll(ctx context.Context, filter FilterDTO) ([]Event, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Record(ctx context.Context, dto CreateEventDTO) error {
	beforeJSON, err := marshalSnapshot(dto.Before)
	if err != nil {
		return err
	}
	afterJSON, err := marshalSnapshot(dto.After)
	if err != nil {
		return err
	}

	var summary *string
	if dto.Summary != "" {
		summary = &dto.Summary
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO audit_events (
			entity_type, entity_id, action, actor_user_id, actor_username,
			summary, before_json, after_json
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`, dto.EntityType, dto.EntityID, dto.Action, dto.Actor.UserID, dto.Actor.Username, summary, beforeJSON, afterJSON)

	return err
}

func (r *repository) GetAll(ctx context.Context, filter FilterDTO) ([]Event, error) {
	query := `
		SELECT id, entity_type, entity_id, action, actor_user_id, actor_username,
			summary, before_json, after_json, created_at
		FROM audit_events
		WHERE 1 = 1
	`
	args := []any{}
	arg := 1

	if filter.EntityType != nil {
		query += fmt.Sprintf(" AND entity_type = $%d", arg)
		args = append(args, *filter.EntityType)
		arg++
	}
	if filter.EntityID != nil {
		query += fmt.Sprintf(" AND entity_id = $%d", arg)
		args = append(args, *filter.EntityID)
		arg++
	}
	if filter.From != nil {
		query += fmt.Sprintf(" AND datetime(created_at) >= datetime($%d)", arg)
		args = append(args, *filter.From)
		arg++
	}
	if filter.To != nil {
		query += fmt.Sprintf(" AND datetime(created_at) <= datetime($%d)", arg)
		args = append(args, *filter.To)
		arg++
	}

	query += " ORDER BY datetime(created_at) DESC, id DESC;"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []Event{}
	for rows.Next() {
		var event Event
		var actorID sql.NullInt64
		var actorUsername, summary, beforeJSON, afterJSON sql.NullString
		if err := rows.Scan(
			&event.ID, &event.EntityType, &event.EntityID, &event.Action,
			&actorID, &actorUsername, &summary, &beforeJSON, &afterJSON,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}
		if actorID.Valid {
			id := int(actorID.Int64)
			event.ActorUserID = &id
		}
		if actorUsername.Valid {
			event.ActorUsername = &actorUsername.String
		}
		if summary.Valid {
			event.Summary = &summary.String
		}
		if beforeJSON.Valid {
			event.BeforeJSON = &beforeJSON.String
		}
		if afterJSON.Valid {
			event.AfterJSON = &afterJSON.String
		}
		events = append(events, event)
	}

	return events, rows.Err()
}
