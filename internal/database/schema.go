package database

import (
	"fmt"
)

// InitSchema creates the snippets table if it doesn't exist
func InitSchema() error {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS snippets (
	    id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL DEFAULT '',
		code TEXT NOT NULL,
		language TEXT DEFAULT '',
	    tags TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_title ON snippets(title);
	CREATE INDEX IF NOT EXISTS idx_tags ON snippets(tags);
	`

	_, err := DB.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create snippets table: %w", err)
	}
	return nil
}
