-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE name = $1;

-- name: Reset :exec
DELETE FROM users;  

-- name: GetUsers :many
SELECT name FROM users;

-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetFeeds :many
SELECT name, url, user_id FROM feeds;

-- name: GetUsername :one
SELECT name FROM users WHERE id = $1;

-- name: GetFeedName :one
SELECT id FROM feeds WHERE url = $1;

-- name: CreateFeedFollow :one
WITH inserted_feed_follows as (
INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *)
SELECT inserted_feed_follows.*, feeds.name as feed_name, users.name as user_name
FROM ((inserted_feed_follows INNER JOIN users ON users.id = inserted_feed_follows.user_id) 
INNER JOIN feeds ON feeds.id = inserted_feed_follows.feed_id);

-- name: GetFeedFollowsForUser :many
SELECT feed_follows.*, feeds.name as feed_name
FROM ((feed_follows INNER JOIN users ON users.id = feed_follows.user_id) 
INNER JOIN feeds ON feeds.id = feed_follows.feed_id)
WHERE users.name = $1;

-- name: UnfollowFeed :exec
DELETE FROM feed_follows
WHERE user_id = $1 AND feed_id = $2;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = $1, updated_at = $1
WHERE id = $2;

-- name: GetNextFeedToFetch :one
UPDATE feeds
SET last_fetched_at = $1, updated_at = $1
WHERE id = $2;