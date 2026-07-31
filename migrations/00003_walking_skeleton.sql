-- +goose Up
CREATE TABLE sources (
    id uuid PRIMARY KEY,
    key citext NOT NULL UNIQUE,
    name text NOT NULL,
    homepage_url text NOT NULL,
    enabled boolean NOT NULL DEFAULT TRUE,
    last_success_at timestamptz,
    last_error_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE observations (
    id uuid PRIMARY KEY,
    source_id uuid NOT NULL REFERENCES sources(id),
    external_id text NOT NULL,
    version integer NOT NULL CHECK (version > 0),
    content_hash text NOT NULL,
    observed_at timestamptz NOT NULL,
    occurred_at timestamptz,
    location geometry(Point, 4326),
    payload jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (source_id, external_id, version),
    UNIQUE (source_id, external_id, content_hash)
);

CREATE INDEX observations_location_gix ON observations USING gist (location);

CREATE TABLE events (
    id uuid PRIMARY KEY,
    slug citext NOT NULL UNIQUE,
    event_type text NOT NULL CHECK (event_type IN ('earthquake', 'wildfire', 'flood', 'storm', 'volcano', 'other')),
    lifecycle text NOT NULL CHECK (lifecycle IN ('active', 'monitoring', 'resolved')),
    priority text NOT NULL CHECK (priority IN ('advisory', 'moderate', 'high', 'critical')),
    confidence text NOT NULL CHECK (confidence IN ('low', 'moderate', 'high')),
    location_name text NOT NULL,
    location geometry(Point, 4326) NOT NULL,
    started_at timestamptz NOT NULL,
    verified_at timestamptz NOT NULL,
    published_revision_id uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX events_location_gix ON events USING gist (location);
CREATE INDEX events_public_rank_idx ON events (lifecycle, priority, verified_at DESC);

CREATE TABLE event_revisions (
    id uuid PRIMARY KEY,
    event_id uuid NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    revision integer NOT NULL CHECK (revision > 0),
    title text NOT NULL,
    summary text NOT NULL,
    created_by uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    published_at timestamptz,
    UNIQUE (event_id, revision)
);

ALTER TABLE events
    ADD CONSTRAINT events_published_revision_fk
    FOREIGN KEY (published_revision_id) REFERENCES event_revisions(id);

CREATE TABLE event_observations (
    event_id uuid NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    observation_id uuid NOT NULL REFERENCES observations(id),
    linked_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (event_id, observation_id)
);

CREATE TABLE timeline_entries (
    id uuid PRIMARY KEY,
    event_id uuid NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    observation_id uuid REFERENCES observations(id),
    occurred_at timestamptz NOT NULL,
    title text NOT NULL,
    body text NOT NULL,
    published boolean NOT NULL DEFAULT FALSE,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX timeline_entries_event_time_idx ON timeline_entries (event_id, occurred_at DESC);

-- +goose Down
DROP TABLE IF EXISTS timeline_entries;
DROP TABLE IF EXISTS event_observations;
ALTER TABLE events DROP CONSTRAINT IF EXISTS events_published_revision_fk;
DROP TABLE IF EXISTS event_revisions;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS observations;
DROP TABLE IF EXISTS sources;
