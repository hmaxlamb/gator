-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (
        user_id, feed_id
    ) VALUES (
        $1, $2
    )
    RETURNING *
)
SELECT
    inserted_feed_follow.*,
    feeds.name as feed_name,
    users.name as user_name
FROM inserted_feed_follow
INNER JOIN users on inserted_feed_follow.user_id = users.id
INNER JOIN feeds on inserted_feed_follow.feed_id = feeds.id;
