package database

import (
	"fmt"
)

func InitSchema() error {
	fmt.Println("Initializing database schema")
    
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS snippets (
	    id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
		code TEXT NOT NULL,
		language TEXT,
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