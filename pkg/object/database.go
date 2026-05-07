package object

import (
	"database/sql"
	"fmt"
)

const DATABASE_OBJ = "DATABASE"

type Database struct {
	DB *sql.DB
}

func (d *Database) Type() ObjectType { return DATABASE_OBJ }
func (d *Database) Inspect() string {
	if d.DB != nil {
		return fmt.Sprintf("<DatabaseConnection>")
	}
	return "<DatabaseConnection (Closed)>"
}

func (d *Database) HashKey() HashKey {
	// Not usable as a map key normally, but required to satisfy interface if needed
	return HashKey{Type: DATABASE_OBJ, Value: 0} 
}
