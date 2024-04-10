// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: user_data.sql

package taskdb

import (
	"context"
)

const deleteUserTaskData = `-- name: DeleteUserTaskData :exec
DELETE FROM
    task
WHERE
    user_id = ?
`

func (q *Queries) DeleteUserTaskData(ctx context.Context, userID string) error {
	_, err := q.db.ExecContext(ctx, deleteUserTaskData, userID)
	return err
}