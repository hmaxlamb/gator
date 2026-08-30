-- +goose Up
CREATE TABLE feed_follows(
    id int PRIMARY KEY,
    user_id uuid NOT NULL,
    feed_id uuid NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_feed FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE,
    CONSTRAINT unique_user_feed_pair UNIQUE (user_id, feed_id)
);

-- +goose Down
DROP TABLE feed_follows;
