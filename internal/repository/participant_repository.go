package repository

import (
	"context"
	"fmt"

	"meeting-slot-service/internal/database"
	"meeting-slot-service/internal/models"
)

type participantRepository struct {
	db *database.Database
}

// NewParticipantRepository creates a new participant repository
func NewParticipantRepository(db *database.Database) ParticipantRepository {
	return &participantRepository{db: db}
}

func (r *participantRepository) AddParticipant(ctx context.Context, participant *models.EventParticipant) error {
	db, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `INSERT INTO event_participants (event_id, user_id, status, created_at, updated_at) 
			  VALUES (?, ?, ?, NOW(), NOW())`
	result, err := db.ExecContext(ctx, query, participant.EventID, participant.UserID, participant.Status)
	if err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	// Get the inserted ID
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	participant.ID = uint(id)

	return nil
}

func (r *participantRepository) GetEventParticipants(ctx context.Context, eventID string) ([]models.EventParticipant, error) {
	db, err := r.db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `SELECT ep.id, ep.event_id, ep.user_id, ep.status, ep.created_at, ep.updated_at,
			  u.id, u.name, u.email, u.created_at, u.updated_at
			  FROM event_participants ep
			  LEFT JOIN users u ON ep.user_id = u.id
			  WHERE ep.event_id = ?`

	rows, err := db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}
	defer rows.Close()

	var participants []models.EventParticipant
	for rows.Next() {
		var p models.EventParticipant
		var user models.User
		if err := rows.Scan(&p.ID, &p.EventID, &p.UserID, &p.Status, &p.CreatedAt, &p.UpdatedAt,
			&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan participant: %w", err)
		}
		p.User = &user
		participants = append(participants, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return participants, nil
}

func (r *participantRepository) GetParticipant(ctx context.Context, eventID, userID string) (*models.EventParticipant, error) {
	db, err := r.db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `SELECT id, event_id, user_id, status, created_at, updated_at
			  FROM event_participants
			  WHERE event_id = ? AND user_id = ?`

	var p models.EventParticipant
	err = db.QueryRowContext(ctx, query, eventID, userID).Scan(
		&p.ID, &p.EventID, &p.UserID, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("participant not found: %w", err)
	}

	return &p, nil
}

func (r *participantRepository) RemoveParticipant(ctx context.Context, eventID, userID string) error {
	db, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `DELETE FROM event_participants WHERE event_id = ? AND user_id = ?`
	result, err := db.ExecContext(ctx, query, eventID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove participant: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("participant not found")
	}

	return nil
}

func (r *participantRepository) UpdateParticipantStatus(ctx context.Context, eventID, userID, status string) error {
	db, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `UPDATE event_participants SET status = ?, updated_at = NOW() WHERE event_id = ? AND user_id = ?`
	result, err := db.ExecContext(ctx, query, status, eventID, userID)
	if err != nil {
		return fmt.Errorf("failed to update participant status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("participant not found")
	}

	return nil
}
