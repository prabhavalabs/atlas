// Package fixture imports deterministic local source data for development and tests.
package fixture

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const maximumFixtureBytes = 1 << 20

type sourceDefinition struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	HomepageURL string `json:"homepageUrl"`
}

type payload struct {
	Source       sourceDefinition `json:"source"`
	ExternalID   string           `json:"externalId"`
	Slug         string           `json:"slug"`
	Type         string           `json:"type"`
	Lifecycle    string           `json:"lifecycle"`
	Priority     string           `json:"priority"`
	Confidence   string           `json:"confidence"`
	Title        string           `json:"title"`
	Summary      string           `json:"summary"`
	LocationName string           `json:"locationName"`
	Latitude     float64          `json:"latitude"`
	Longitude    float64          `json:"longitude"`
	ObservedAt   time.Time        `json:"observedAt"`
	StartedAt    time.Time        `json:"startedAt"`
	VerifiedAt   time.Time        `json:"verifiedAt"`
}

// Result describes whether importing a fixture created a new observation revision.
type Result struct {
	EventID string
	Changed bool
}

// Import validates and transactionally imports one fixture event.
func Import(ctx context.Context, pool *pgxpool.Pool, reader io.Reader) (Result, error) {
	raw, err := io.ReadAll(io.LimitReader(reader, maximumFixtureBytes+1))
	if err != nil {
		return Result{}, fmt.Errorf("read fixture: %w", err)
	}
	if len(raw) > maximumFixtureBytes {
		return Result{}, errors.New("fixture exceeds 1 MiB limit")
	}

	var fixture payload
	if err := json.Unmarshal(raw, &fixture); err != nil {
		return Result{}, fmt.Errorf("decode fixture: %w", err)
	}
	if err := fixture.validate(); err != nil {
		return Result{}, err
	}
	canonical, err := json.Marshal(fixture)
	if err != nil {
		return Result{}, fmt.Errorf("canonicalize fixture: %w", err)
	}
	contentDigest := sha256.Sum256(canonical)
	contentHash := hex.EncodeToString(contentDigest[:])

	transaction, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Result{}, fmt.Errorf("begin fixture import: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	sourceID := uuid.New()
	err = transaction.QueryRow(ctx, `
		INSERT INTO sources (id, key, name, homepage_url, last_success_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (key) DO UPDATE SET
			name = EXCLUDED.name,
			homepage_url = EXCLUDED.homepage_url,
			last_success_at = EXCLUDED.last_success_at,
			updated_at = now()
		RETURNING id
	`, sourceID, fixture.Source.Key, fixture.Source.Name, fixture.Source.HomepageURL, fixture.ObservedAt).Scan(&sourceID)
	if err != nil {
		return Result{}, fmt.Errorf("upsert fixture source: %w", err)
	}

	var existingEventID uuid.UUID
	err = transaction.QueryRow(ctx, `
		SELECT eo.event_id
		FROM observations o
		JOIN event_observations eo ON eo.observation_id = o.id
		WHERE o.source_id = $1 AND o.external_id = $2 AND o.content_hash = $3
	`, sourceID, fixture.ExternalID, contentHash).Scan(&existingEventID)
	if err == nil {
		if err := transaction.Commit(ctx); err != nil {
			return Result{}, fmt.Errorf("commit unchanged fixture: %w", err)
		}
		return Result{EventID: existingEventID.String(), Changed: false}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Result{}, fmt.Errorf("check fixture idempotency: %w", err)
	}

	var observationVersion int
	err = transaction.QueryRow(ctx, `
		SELECT COALESCE(MAX(version), 0) + 1
		FROM observations
		WHERE source_id = $1 AND external_id = $2
	`, sourceID, fixture.ExternalID).Scan(&observationVersion)
	if err != nil {
		return Result{}, fmt.Errorf("select observation version: %w", err)
	}

	observationID := uuid.New()
	_, err = transaction.Exec(ctx, `
		INSERT INTO observations (
			id, source_id, external_id, version, content_hash, observed_at,
			occurred_at, location, payload
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			ST_SetSRID(ST_MakePoint($8, $9), 4326), $10
		)
	`, observationID, sourceID, fixture.ExternalID, observationVersion, contentHash,
		fixture.ObservedAt, fixture.StartedAt, fixture.Longitude, fixture.Latitude, canonical)
	if err != nil {
		return Result{}, fmt.Errorf("insert observation: %w", err)
	}

	eventID, revision, err := upsertEvent(ctx, transaction, fixture)
	if err != nil {
		return Result{}, err
	}
	_, err = transaction.Exec(ctx, `
		INSERT INTO event_observations (event_id, observation_id)
		VALUES ($1, $2)
	`, eventID, observationID)
	if err != nil {
		return Result{}, fmt.Errorf("link observation to event: %w", err)
	}
	_, err = transaction.Exec(ctx, `
		INSERT INTO timeline_entries (
			id, event_id, observation_id, occurred_at, title, body, published
		) VALUES ($1, $2, $3, $4, $5, $6, TRUE)
	`, uuid.New(), eventID, observationID, fixture.StartedAt, fixture.Title, fixture.Summary)
	if err != nil {
		return Result{}, fmt.Errorf("insert fixture timeline entry: %w", err)
	}

	if err := transaction.Commit(ctx); err != nil {
		return Result{}, fmt.Errorf("commit fixture import: %w", err)
	}
	return Result{EventID: eventID.String(), Changed: revision > 0}, nil
}

func upsertEvent(ctx context.Context, transaction pgx.Tx, fixture payload) (uuid.UUID, int, error) {
	var eventID uuid.UUID
	var currentRevision int
	err := transaction.QueryRow(ctx, `
		SELECT e.id, COALESCE((
			SELECT MAX(r.revision)
			FROM event_revisions r
			WHERE r.event_id = e.id
		), 0)::integer
		FROM events e
		WHERE e.slug = $1
		FOR UPDATE
	`, fixture.Slug).Scan(&eventID, &currentRevision)
	if errors.Is(err, pgx.ErrNoRows) {
		eventID = uuid.New()
		_, err = transaction.Exec(ctx, `
			INSERT INTO events (
				id, slug, event_type, lifecycle, priority, confidence,
				location_name, location, started_at, verified_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7,
				ST_SetSRID(ST_MakePoint($8, $9), 4326), $10, $11
			)
		`, eventID, fixture.Slug, fixture.Type, fixture.Lifecycle, fixture.Priority,
			fixture.Confidence, fixture.LocationName, fixture.Longitude, fixture.Latitude,
			fixture.StartedAt, fixture.VerifiedAt)
		if err != nil {
			return uuid.Nil, 0, fmt.Errorf("insert event: %w", err)
		}
		currentRevision = 0
	} else if err != nil {
		return uuid.Nil, 0, fmt.Errorf("select event revision: %w", err)
	}

	nextRevision := currentRevision + 1
	revisionID := uuid.New()
	_, err = transaction.Exec(ctx, `
		INSERT INTO event_revisions (
			id, event_id, revision, title, summary, published_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`, revisionID, eventID, nextRevision, fixture.Title, fixture.Summary, fixture.VerifiedAt)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("insert event revision: %w", err)
	}
	_, err = transaction.Exec(ctx, `
		UPDATE events SET
			event_type = $2,
			lifecycle = $3,
			priority = $4,
			confidence = $5,
			location_name = $6,
			location = ST_SetSRID(ST_MakePoint($7, $8), 4326),
			started_at = $9,
			verified_at = $10,
			published_revision_id = $11,
			updated_at = now()
		WHERE id = $1
	`, eventID, fixture.Type, fixture.Lifecycle, fixture.Priority, fixture.Confidence,
		fixture.LocationName, fixture.Longitude, fixture.Latitude, fixture.StartedAt,
		fixture.VerifiedAt, revisionID)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("publish event revision: %w", err)
	}
	return eventID, nextRevision, nil
}

func (fixture payload) validate() error {
	if fixture.Source.Key == "" || fixture.Source.Name == "" || fixture.Source.HomepageURL == "" {
		return errors.New("fixture source key, name, and homepageUrl are required")
	}
	if fixture.ExternalID == "" || fixture.Slug == "" || fixture.Title == "" || fixture.Summary == "" {
		return errors.New("fixture externalId, slug, title, and summary are required")
	}
	if fixture.Latitude < -90 || fixture.Latitude > 90 || fixture.Longitude < -180 || fixture.Longitude > 180 {
		return errors.New("fixture coordinates are outside valid bounds")
	}
	if fixture.ObservedAt.IsZero() || fixture.StartedAt.IsZero() || fixture.VerifiedAt.IsZero() {
		return errors.New("fixture observedAt, startedAt, and verifiedAt are required")
	}
	return nil
}
