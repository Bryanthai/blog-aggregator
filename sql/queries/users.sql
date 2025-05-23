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
SET last_fetch_at = $1, updated_at = $1
WHERE url = $2;

-- name: GetNextFeedToFetch :one
SELECT url FROM feeds ORDER BY last_fetch_at ASC NULLS FIRST;

-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: GetPostsByUser :many
SELECT posts.* 
FROM posts INNER JOIN feed_follows ON posts.feed_id = feed_follows.feed_id
WHERE feed_follows.user_id = $1 AND posts.feed_id = $2
ORDER BY published_at DESC LIMIT $3;

-- name: CheckPostByURL :one
SELECT * FROM posts WHERE URL = $1; 

-- name: UpdatePost :exec
UPDATE posts
SET updated_at = $1, title = $2, description = $3, published_at = $4
WHERE url = $5;