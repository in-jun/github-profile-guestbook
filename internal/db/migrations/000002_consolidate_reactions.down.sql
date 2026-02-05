CREATE TABLE likes (
    id         BIGSERIAL PRIMARY KEY,
    message_id BIGINT    NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id    BIGINT    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT uq_like UNIQUE (message_id, user_id)
);

CREATE TABLE dislikes (
    id         BIGSERIAL PRIMARY KEY,
    message_id BIGINT    NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id    BIGINT    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT uq_dislike UNIQUE (message_id, user_id)
);

INSERT INTO likes (message_id, user_id)
SELECT message_id, user_id FROM reactions WHERE type = 1;

INSERT INTO dislikes (message_id, user_id)
SELECT message_id, user_id FROM reactions WHERE type = -1;

CREATE INDEX idx_likes_message ON likes (message_id);
CREATE INDEX idx_dislikes_message ON dislikes (message_id);

DROP TABLE reactions;

ALTER TABLE users DROP CONSTRAINT uq_users_github_login;
