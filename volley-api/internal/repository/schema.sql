-- Enable PostGIS extension
CREATE EXTENSION IF NOT EXISTS postgis;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Refresh tokens table for managing long-lived authentication sessions
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE, -- SHA-256 hash of the refresh token
    device_info VARCHAR(255), -- Optional: device/client identifier
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ -- NULL if active, set when revoked
);

-- Indexes for refresh tokens
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- Games table
CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category VARCHAR(50) NOT NULL,
    title VARCHAR(255),
    description TEXT,

    -- Location (using PostGIS)
    location_name VARCHAR(255) NOT NULL,
    location_address VARCHAR(255),
    location_point geography(Point, 4326), -- WGS 84 coordinate system
    location_notes TEXT,

    -- Game details
    start_time TIMESTAMPTZ NOT NULL,
    duration_minutes INTEGER NOT NULL,
    max_participants INTEGER NOT NULL,

    -- Pricing (embedded)
    pricing_type VARCHAR(50) NOT NULL,
    pricing_amount_cents INTEGER NOT NULL DEFAULT 0,
    pricing_currency VARCHAR(3) NOT NULL DEFAULT 'USD',

    signup_deadline TIMESTAMPTZ NOT NULL,
    drop_deadline TIMESTAMPTZ, -- Optional deadline for dropping from game
    skill_level VARCHAR(50) NOT NULL DEFAULT 'all',
    notes TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'open',
    cancelled_at TIMESTAMPTZ, -- When the game was cancelled (NULL if not cancelled)

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_games_start_time ON games(start_time) where status = 'open';
CREATE INDEX IF NOT EXISTS idx_games_category_status ON games(category, status);
CREATE INDEX IF NOT EXISTS idx_games_owner_id ON games(owner_id);

-- Spatial index for location-based queries
CREATE INDEX IF NOT EXISTS idx_games_location_point ON games USING GIST(location_point);

-- Teams table for games that support team-based play
CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7), -- Hex color code
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Participants table to track user participation in games
CREATE TABLE IF NOT EXISTS participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team_id UUID REFERENCES teams(id) ON DELETE SET NULL,

    -- Participant status
    status VARCHAR(50) NOT NULL DEFAULT 'confirmed', -- confirmed, waitlist, declined, removed

    -- Payment tracking
    paid BOOLEAN NOT NULL DEFAULT FALSE,
    payment_amount_cents INTEGER,

    -- Additional info
    notes TEXT,

    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure a user can only participate once in a game
    UNIQUE(game_id, user_id)
);

-- Indexes for teams
CREATE INDEX IF NOT EXISTS idx_teams_game_id ON teams(game_id);

-- Indexes for participants
CREATE INDEX IF NOT EXISTS idx_participants_game_id ON participants(game_id);
CREATE INDEX IF NOT EXISTS idx_participants_user_id ON participants(user_id);
CREATE INDEX IF NOT EXISTS idx_participants_status ON participants(game_id, status);
