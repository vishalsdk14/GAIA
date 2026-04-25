// Copyright 2026 GAIA Contributors
//
// Licensed under the MIT License.
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://opensource.org/licenses/MIT
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package state

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gaia/kernel/pkg/types"
)

// TaskStore implements Tier 2 (Task History/Warm State).
// It manages the persistence of Task records, allowing for resumption after kernel restarts.
type TaskStore struct {
	db *sql.DB
}

// NewTaskStore initializes the TaskStore with an existing SQLite connection.
func NewTaskStore(db *sql.DB) (*TaskStore, error) {
	query := `
	CREATE TABLE IF NOT EXISTS tasks (
		task_id TEXT PRIMARY KEY,
		goal TEXT NOT NULL,
		status TEXT NOT NULL,
		task_data JSON NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(query); err != nil {
		return nil, fmt.Errorf("state: failed to create tasks table: %w", err)
	}
	return &TaskStore{db: db}, nil
}

// SaveTask checkpoints the full state of a task.
func (s *TaskStore) SaveTask(task *types.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	query := `
	INSERT INTO tasks (task_id, goal, status, task_data, updated_at)
	VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(task_id) DO UPDATE SET
		status=excluded.status,
		task_data=excluded.task_data,
		updated_at=CURRENT_TIMESTAMP;`

	_, err = s.db.Exec(query, task.TaskID, task.Goal, string(task.Status), string(data))
	return err
}

// GetTask retrieves a task by ID for resumption.
func (s *TaskStore) GetTask(taskID string) (*types.Task, error) {
	var dataStr string
	err := s.db.QueryRow(`SELECT task_data FROM tasks WHERE task_id = ?`, taskID).Scan(&dataStr)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var task types.Task
	if err := json.Unmarshal([]byte(dataStr), &task); err != nil {
		return nil, err
	}
	return &task, nil
}
