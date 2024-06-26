// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: auth.sql

package authdb

import (
	"context"
	"time"
)

const createAuth = `-- name: CreateAuth :one
INSERT INTO
    auth (user_id, username, hashed_password)
VALUES
    (?, ?, ?) RETURNING user_id, username, hashed_password, created_at, updated_at
`

type CreateAuthParams struct {
	UserID         string `json:"user_id"`
	Username       string `json:"username"`
	HashedPassword string `json:"hashed_password"`
}

func (q *Queries) CreateAuth(ctx context.Context, arg CreateAuthParams) (Auth, error) {
	row := q.db.QueryRowContext(ctx, createAuth, arg.UserID, arg.Username, arg.HashedPassword)
	var i Auth
	err := row.Scan(
		&i.UserID,
		&i.Username,
		&i.HashedPassword,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteUser = `-- name: DeleteUser :exec
DELETE FROM
    auth
WHERE
    user_id = ?
`

func (q *Queries) DeleteUser(ctx context.Context, userID string) error {
	_, err := q.db.ExecContext(ctx, deleteUser, userID)
	return err
}

const findUserByUsername = `-- name: FindUserByUsername :one
SELECT
    user_id, username, hashed_password, created_at, updated_at
FROM
    auth
WHERE
    username = ?
`

func (q *Queries) FindUserByUsername(ctx context.Context, username string) (Auth, error) {
	row := q.db.QueryRowContext(ctx, findUserByUsername, username)
	var i Auth
	err := row.Scan(
		&i.UserID,
		&i.Username,
		&i.HashedPassword,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateUser = `-- name: UpdateUser :one
UPDATE
    auth
SET
    username = ?,
    hashed_password = ?,
    updated_at = ?
WHERE
    user_id = ? RETURNING user_id, username, hashed_password, created_at, updated_at
`

type UpdateUserParams struct {
	Username       string    `json:"username"`
	HashedPassword string    `json:"hashed_password"`
	UpdatedAt      time.Time `json:"updated_at"`
	UserID         string    `json:"user_id"`
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (Auth, error) {
	row := q.db.QueryRowContext(ctx, updateUser,
		arg.Username,
		arg.HashedPassword,
		arg.UpdatedAt,
		arg.UserID,
	)
	var i Auth
	err := row.Scan(
		&i.UserID,
		&i.Username,
		&i.HashedPassword,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
