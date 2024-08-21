CREATE TABLE users (
    id           BIGSERIAL   PRIMARY KEY,
    github_id    BIGINT      NOT NULL UNIQUE,
    github_login TEXT        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE comments (
    id             BIGSERIAL   PRIMARY KEY,
    receiver_id    BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    author_id      BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content        TEXT        NOT NULL,
    is_owner_liked BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_comment_receiver_author UNIQUE (receiver_id, author_id)
);

CREATE TABLE likes (
    id         BIGSERIAL PRIMARY KEY,
    comment_id BIGINT    NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    user_id    BIGINT    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT uq_like UNIQUE (comment_id, user_id)
);

CREATE TABLE dislikes (
    id         BIGSERIAL PRIMARY KEY,
    comment_id BIGINT    NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    user_id    BIGINT    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT uq_dislike UNIQUE (comment_id, user_id)
);

CREATE TABLE refresh_tokens (
    id         BIGSERIAL   PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comments_receiver      ON comments       (receiver_id);
CREATE INDEX idx_comments_author        ON comments       (author_id);
CREATE INDEX idx_likes_comment          ON likes          (comment_id);
CREATE INDEX idx_dislikes_comment       ON dislikes       (comment_id);
CREATE INDEX idx_refresh_tokens_user    ON refresh_tokens (user_id);
CREATE INDEX idx_refresh_tokens_hash    ON refresh_tokens (token_hash);
