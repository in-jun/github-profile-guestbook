ALTER TABLE users ADD CONSTRAINT uq_users_github_login UNIQUE (github_login);

CREATE TABLE reactions (
    message_id BIGINT   NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id    BIGINT   NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type       SMALLINT NOT NULL CHECK (type IN (1, -1)),
    PRIMARY KEY (message_id, user_id)
);

INSERT INTO reactions (message_id, user_id, type)
SELECT message_id, user_id, 1 FROM likes;

INSERT INTO reactions (message_id, user_id, type)
SELECT message_id, user_id, -1 FROM dislikes
ON CONFLICT (message_id, user_id) DO NOTHING;

DROP TABLE likes;
DROP TABLE dislikes;
