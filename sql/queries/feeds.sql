-- name: CreateFeed :one
INSERT INTO feeds(id, created_at, updated_at, name, url, user_id)
VALUES(
	$1,
	$2,
	$3,
	$4,
	$5,
	$6
)
RETURNING *;


-- name: GetFeed :one
SELECT * FROM feeds
WHERE url = $1;

-- name: MarkFeedFetch :exec
UPDATE feeds
SET updated_at = $1,
    last_fetched_at = $1
WHERE feeds.id = $2;

-- name: GetNextFeedToFetch :one
SELECT feeds.* 
FROM feeds
INNER JOIN feed_follows on feeds.id = feed_follows.feed_id
WHERE feed_follows.user_id = $1
ORDER BY feeds.last_fetched_at ASC NULLS FIRST
LIMIT 1;
