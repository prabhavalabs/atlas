package identity

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Role is an administrator authorization level.
type Role string

const (
	// RoleAdministrator can manage identities and all editorial content.
	RoleAdministrator Role = "administrator"
	// RoleEditor can review and edit editorial content.
	RoleEditor Role = "editor"
	// RoleViewer has read-only administration access.
	RoleViewer Role = "viewer"
)

// ErrInvalidSession deliberately hides whether a token is unknown, revoked, or expired.
var ErrInvalidSession = errors.New("invalid session")

// ErrInvalidCredentials deliberately hides which login field was wrong.
var ErrInvalidCredentials = errors.New("invalid email or password")

// User is the safe administrator identity exposed to handlers.
type User struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	Role        Role      `json:"role"`
	Active      bool      `json:"active"`
}

// CredentialUser is used only inside login verification.
type CredentialUser struct {
	User
	PasswordHash string
}

// Session is a persisted administrator session.
type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
}

// AuthenticatedSession combines a valid session, its user, and CSRF hash.
type AuthenticatedSession struct {
	SessionID uuid.UUID
	User      User
	CSRFHash  []byte
	ExpiresAt time.Time
}

// Repository stores administrator identities and sessions.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a Postgres-backed identity repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// CreateUser creates one local administrator identity.
func (repository *Repository) CreateUser(ctx context.Context, email string, displayName string, passwordHash string, role Role) (User, error) {
	if !validRole(role) {
		return User{}, errors.New("role must be administrator, editor, or viewer")
	}
	user := User{ID: uuid.New(), Email: email, DisplayName: displayName, Role: role, Active: true}
	err := repository.pool.QueryRow(ctx, `
		INSERT INTO users (id, email, display_name, password_hash, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING email::text, display_name, role, active
	`, user.ID, email, displayName, passwordHash, role).Scan(&user.Email, &user.DisplayName, &user.Role, &user.Active)
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

// FindCredentialUserByEmail returns credentials for constant-response login verification.
func (repository *Repository) FindCredentialUserByEmail(ctx context.Context, email string) (CredentialUser, error) {
	var user CredentialUser
	err := repository.pool.QueryRow(ctx, `
		SELECT id, email::text, display_name, role, active, password_hash
		FROM users
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &user.DisplayName, &user.Role, &user.Active, &user.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return CredentialUser{}, ErrInvalidCredentials
	}
	if err != nil {
		return CredentialUser{}, fmt.Errorf("find credential user: %w", err)
	}
	return user, nil
}

// CreateSession persists only one-way hashes of browser-facing secrets.
func (repository *Repository) CreateSession(ctx context.Context, userID uuid.UUID, tokens SessionTokens, expiresAt time.Time, now time.Time) (Session, error) {
	session := Session{ID: uuid.New(), UserID: userID, ExpiresAt: expiresAt}
	_, err := repository.pool.Exec(ctx, `
		INSERT INTO sessions (
			id, user_id, token_hash, csrf_hash, expires_at, last_seen_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`, session.ID, userID, tokens.SessionHash, tokens.CSRFHash, expiresAt, now)
	if err != nil {
		return Session{}, fmt.Errorf("create session: %w", err)
	}
	return session, nil
}

// AuthenticateSession returns a session only when its user and token remain valid.
func (repository *Repository) AuthenticateSession(ctx context.Context, rawToken string, now time.Time) (AuthenticatedSession, error) {
	tokenHash := sha256.Sum256([]byte(rawToken))
	var authenticated AuthenticatedSession
	err := repository.pool.QueryRow(ctx, `
		SELECT
			s.id,
			s.expires_at,
			s.csrf_hash,
			u.id,
			u.email::text,
			u.display_name,
			u.role,
			u.active
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1
		  AND s.revoked_at IS NULL
		  AND s.expires_at > $2
		  AND u.active = TRUE
	`, tokenHash[:], now).Scan(
		&authenticated.SessionID,
		&authenticated.ExpiresAt,
		&authenticated.CSRFHash,
		&authenticated.User.ID,
		&authenticated.User.Email,
		&authenticated.User.DisplayName,
		&authenticated.User.Role,
		&authenticated.User.Active,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return AuthenticatedSession{}, ErrInvalidSession
	}
	if err != nil {
		return AuthenticatedSession{}, fmt.Errorf("authenticate session: %w", err)
	}
	return authenticated, nil
}

// RevokeSession invalidates one session without deleting its audit history.
func (repository *Repository) RevokeSession(ctx context.Context, sessionID uuid.UUID, revokedAt time.Time) error {
	result, err := repository.pool.Exec(ctx, `
		UPDATE sessions
		SET revoked_at = $2
		WHERE id = $1 AND revoked_at IS NULL
	`, sessionID, revokedAt)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrInvalidSession
	}
	return nil
}

func validRole(role Role) bool {
	return role == RoleAdministrator || role == RoleEditor || role == RoleViewer
}
