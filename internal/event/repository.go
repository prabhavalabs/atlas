// Package event owns canonical disaster event read and editorial write behavior.
package event

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound indicates that a published event does not exist.
var ErrNotFound = errors.New("published event not found")

// ErrRevisionConflict indicates an editorial write based on stale event data.
var ErrRevisionConflict = errors.New("event revision conflict")

// PublicEvent is the stable anonymous event representation.
type PublicEvent struct {
	ID           string    `json:"id"`
	Slug         string    `json:"slug"`
	Type         string    `json:"type"`
	Lifecycle    string    `json:"lifecycle"`
	Priority     string    `json:"priority"`
	Confidence   string    `json:"confidence"`
	Title        string    `json:"title"`
	Summary      string    `json:"summary"`
	LocationName string    `json:"locationName"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	StartedAt    time.Time `json:"startedAt"`
	VerifiedAt   time.Time `json:"verifiedAt"`
	SourceCount  int       `json:"sourceCount"`
	Revision     int       `json:"revision"`
}

// Repository reads published events from Postgres.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates an event repository backed by pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// ListPublic returns only events that have a published revision.
func (repository *Repository) ListPublic(ctx context.Context) ([]PublicEvent, error) {
	rows, err := repository.pool.Query(ctx, publicEventSelect+`
		ORDER BY CASE e.priority
			WHEN 'critical' THEN 1
			WHEN 'high' THEN 2
			WHEN 'moderate' THEN 3
			ELSE 4
		END, e.verified_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list public events: %w", err)
	}
	defer rows.Close()

	events := make([]PublicEvent, 0)
	for rows.Next() {
		publicEvent, scanErr := scanPublicEvent(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		events = append(events, publicEvent)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate public events: %w", err)
	}
	return events, nil
}

// GetPublicBySlug returns one published event by its stable public slug.
func (repository *Repository) GetPublicBySlug(ctx context.Context, slug string) (PublicEvent, error) {
	row := repository.pool.QueryRow(ctx, publicEventSelect+" WHERE e.slug = $1", slug)
	publicEvent, err := scanPublicEvent(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return PublicEvent{}, ErrNotFound
	}
	if err != nil {
		return PublicEvent{}, err
	}
	return publicEvent, nil
}

// UpdateTitleInput contains the optimistic concurrency and audit context for one title edit.
type UpdateTitleInput struct {
	EventID          uuid.UUID
	Title            string
	Reason           string
	ExpectedRevision int
	ActorID          uuid.UUID
	RequestID        string
	Now              time.Time
}

// UpdatePublishedTitle creates and immediately publishes a revision with one audited title change.
func (repository *Repository) UpdatePublishedTitle(ctx context.Context, input UpdateTitleInput) (PublicEvent, error) {
	transaction, err := repository.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return PublicEvent{}, fmt.Errorf("begin title update: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	var slug string
	var currentRevision int
	var currentTitle string
	var summary string
	err = transaction.QueryRow(ctx, `
		SELECT e.slug::text, r.revision, r.title, r.summary
		FROM events e
		JOIN event_revisions r ON r.id = e.published_revision_id
		WHERE e.id = $1
		FOR UPDATE OF e
	`, input.EventID).Scan(&slug, &currentRevision, &currentTitle, &summary)
	if errors.Is(err, pgx.ErrNoRows) {
		return PublicEvent{}, ErrNotFound
	}
	if err != nil {
		return PublicEvent{}, fmt.Errorf("select published event for update: %w", err)
	}
	if currentRevision != input.ExpectedRevision {
		return PublicEvent{}, ErrRevisionConflict
	}

	nextRevision := currentRevision + 1
	revisionID := uuid.New()
	_, err = transaction.Exec(ctx, `
		INSERT INTO event_revisions (
			id, event_id, revision, title, summary, created_by, created_at, published_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
	`, revisionID, input.EventID, nextRevision, input.Title, summary, input.ActorID, input.Now)
	if err != nil {
		return PublicEvent{}, fmt.Errorf("insert title revision: %w", err)
	}
	_, err = transaction.Exec(ctx, `
		UPDATE events
		SET published_revision_id = $2, updated_at = $3
		WHERE id = $1
	`, input.EventID, revisionID, input.Now)
	if err != nil {
		return PublicEvent{}, fmt.Errorf("publish title revision: %w", err)
	}

	beforeHash := contentHash(fmt.Sprintf("%d:%s", currentRevision, currentTitle))
	afterHash := contentHash(fmt.Sprintf("%d:%s", nextRevision, input.Title))
	_, err = transaction.Exec(ctx, `
		INSERT INTO audit_log (
			actor_user_id, action, target_type, target_id, reason,
			before_hash, after_hash, request_id, created_at
		) VALUES ($1, 'event.title.updated', 'event', $2, $3, $4, $5, $6, $7)
	`, input.ActorID, input.EventID.String(), input.Reason, beforeHash, afterHash, input.RequestID, input.Now)
	if err != nil {
		return PublicEvent{}, fmt.Errorf("insert title audit record: %w", err)
	}
	if err := transaction.Commit(ctx); err != nil {
		return PublicEvent{}, fmt.Errorf("commit title update: %w", err)
	}
	return repository.GetPublicBySlug(ctx, slug)
}

const publicEventSelect = `
	SELECT
		e.id::text,
		e.slug::text,
		e.event_type,
		e.lifecycle,
		e.priority,
		e.confidence,
		r.title,
		r.summary,
		e.location_name,
		ST_Y(e.location)::double precision AS latitude,
		ST_X(e.location)::double precision AS longitude,
		e.started_at,
		e.verified_at,
		(SELECT COUNT(DISTINCT eo.observation_id)::integer
		 FROM event_observations eo
		 WHERE eo.event_id = e.id) AS source_count,
		r.revision
	FROM events e
	JOIN event_revisions r ON r.id = e.published_revision_id
`

type scanner interface {
	Scan(dest ...any) error
}

func scanPublicEvent(row scanner) (PublicEvent, error) {
	var publicEvent PublicEvent
	err := row.Scan(
		&publicEvent.ID,
		&publicEvent.Slug,
		&publicEvent.Type,
		&publicEvent.Lifecycle,
		&publicEvent.Priority,
		&publicEvent.Confidence,
		&publicEvent.Title,
		&publicEvent.Summary,
		&publicEvent.LocationName,
		&publicEvent.Latitude,
		&publicEvent.Longitude,
		&publicEvent.StartedAt,
		&publicEvent.VerifiedAt,
		&publicEvent.SourceCount,
		&publicEvent.Revision,
	)
	if err != nil {
		return PublicEvent{}, fmt.Errorf("scan public event: %w", err)
	}
	return publicEvent, nil
}

func contentHash(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
