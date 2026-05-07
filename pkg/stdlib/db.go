//go:build !runner || runner_sqlite

package stdlib

import (
	"database/sql"
	"fmt"
	"jabline/pkg/object"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

var DBBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"open", &object.Builtin{Fn: dbOpen}},
	{"query", &object.Builtin{Fn: dbQuery}},
	{"exec", &object.Builtin{Fn: dbExec}},
	{"close", &object.Builtin{Fn: dbClose}},
	{"db_all", &object.Builtin{Fn: dbAll}},
	{"db_get", &object.Builtin{Fn: dbGet}},
	{"db_run", &object.Builtin{Fn: dbRun}},
	{"db_insert", &object.Builtin{Fn: dbInsert}},
}

func init() {
	NativeModuleRegistry["_db"] = DBBuiltins
	
	// Register global helpers
	Registry = append(Registry, []struct {
		Name   string
		Object object.Object
	}{
		{"db_open", &object.Builtin{Fn: dbOpen}},
		{"db_all", &object.Builtin{Fn: dbAll}},
		{"db_get", &object.Builtin{Fn: dbGet}},
		{"db_run", &object.Builtin{Fn: dbRun}},
		{"db_insert", &object.Builtin{Fn: dbInsert}},
	}...)
}

func dbOpen(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("db.open expects exactly 2 arguments (driver, dsn)")
	}

	driver, ok1 := args[0].(*object.String)
	dsn, ok2 := args[1].(*object.String)

	if !ok1 || !ok2 {
		return newError("db.open expects 2 STRING arguments")
	}

	db, err := sql.Open(driver.Value, dsn.Value)
	if err != nil {
		return newError("failed to open database: %s", err.Error())
	}

	// Ping to verify connection
	if err := db.Ping(); err != nil {
		return newError("failed to connect to database: %s", err.Error())
	}

	return &object.Database{DB: db}
}

func dbQuery(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("db.query expects at least 2 arguments (conn, query)")
	}

	conn, ok := args[0].(*object.Database)
	if !ok {
		return newError("db.query first argument must be a Database connection")
	}
	
	query, ok := args[1].(*object.String)
	if !ok {
		return newError("db.query second argument must be a STRING query")
	}

	// Convert args
	var queryArgs []interface{}
	if len(args) == 3 {
		if arr, isArr := args[2].(*object.Array); isArr {
			queryArgs = make([]interface{}, len(arr.Elements))
			for i, el := range arr.Elements {
				queryArgs[i] = convertToNative(el)
			}
		} else if _, isNull := args[2].(*object.Null); !isNull {
			queryArgs = append(queryArgs, convertToNative(args[2]))
		}
	} else if len(args) > 3 {
		queryArgs = make([]interface{}, len(args)-2)
		for i := 2; i < len(args); i++ {
			queryArgs[i-2] = convertToNative(args[i])
		}
	}

	rows, err := conn.DB.Query(query.Value, queryArgs...)
	if err != nil {
		return newError("query failed: %s", err.Error())
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return newError("failed to get columns: %s", err.Error())
	}

	var results []object.Object

	for rows.Next() {
		// Scan dynamically
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return newError("failed to scan row: %s", err.Error())
		}

		rowHash := make(map[object.HashKey]object.HashPair)
		for i, colName := range cols {
			val := columns[i]
			keyObj := &object.String{Value: colName}
			valObj := convertToJabline(val)
			rowHash[keyObj.HashKey()] = object.HashPair{Key: keyObj, Value: valObj}
		}

		results = append(results, &object.Hash{Pairs: rowHash})
	}

	return &object.Array{Elements: results}
}

func dbExec(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("db.exec expects at least 2 arguments (conn, query)")
	}

	conn, ok := args[0].(*object.Database)
	if !ok {
		return newError("db.exec first argument must be a Database connection")
	}
	
	query, ok := args[1].(*object.String)
	if !ok {
		return newError("db.exec second argument must be a STRING query")
	}

	// Convert args
	var queryArgs []interface{}
	if len(args) == 3 {
		if arr, isArr := args[2].(*object.Array); isArr {
			queryArgs = make([]interface{}, len(arr.Elements))
			for i, el := range arr.Elements {
				queryArgs[i] = convertToNative(el)
			}
		} else if _, isNull := args[2].(*object.Null); !isNull {
			queryArgs = append(queryArgs, convertToNative(args[2]))
		}
	} else if len(args) > 3 {
		queryArgs = make([]interface{}, len(args)-2)
		for i := 2; i < len(args); i++ {
			queryArgs[i-2] = convertToNative(args[i])
		}
	}

	res, err := conn.DB.Exec(query.Value, queryArgs...)
	if err != nil {
		return newError("exec failed: %s", err.Error())
	}

	affected, _ := res.RowsAffected()
	lastInsertId, _ := res.LastInsertId()

	hash := make(map[object.HashKey]object.HashPair)
	
	k1 := &object.String{Value: "rowsAffected"}
	v1 := &object.Integer{Value: affected}
	hash[k1.HashKey()] = object.HashPair{Key: k1, Value: v1}

	k2 := &object.String{Value: "lastInsertId"}
	v2 := &object.Integer{Value: lastInsertId}
	hash[k2.HashKey()] = object.HashPair{Key: k2, Value: v2}

	return &object.Hash{Pairs: hash}
}

func dbClose(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("db.close expects 1 argument (conn)")
	}
	conn, ok := args[0].(*object.Database)
	if !ok {
		return newError("db.close argument must be a Database connection")
	}
	
	if err := conn.DB.Close(); err != nil {
		return newError("failed to close database: %s", err.Error())
	}
	return &object.Null{}
}

// Helpers
func convertToNative(obj object.Object) interface{} {
	switch o := obj.(type) {
	case *object.Integer: return o.Value
	case *object.Float: return o.Value
	case *object.String: return o.Value
	case *object.Boolean: return o.Value
	case *object.Null: return nil
	default: return o.Inspect()
	}
}

func convertToJabline(val interface{}) object.Object {
	switch v := val.(type) {
	case nil: return &object.Null{}
	case int64: return &object.Integer{Value: v}
	case int: return &object.Integer{Value: int64(v)}
	case float64: return &object.Float{Value: v}
	case bool: return &object.Boolean{Value: v}
	case string: return &object.String{Value: v}
	case []byte: return &object.String{Value: string(v)}
	default: return &object.String{Value: fmt.Sprintf("%v", v)}
	}
}

// dbAll(conn, sql, [args]) → Array of rows
func dbAll(args ...object.Object) object.Object {
	return dbQuery(args...)
}

// dbGet(conn, sql, [args]) → First row or null
func dbGet(args ...object.Object) object.Object {
	result := dbQuery(args...)
	if arr, ok := result.(*object.Array); ok {
		if len(arr.Elements) > 0 {
			return arr.Elements[0]
		}
		return &object.Null{}
	}
	return result
}

// dbRun(conn, sql, [args]) → { rowsAffected, lastInsertId }
func dbRun(args ...object.Object) object.Object {
	return dbExec(args...)
}

// dbInsert(conn, table, dataHash) → { rowsAffected, lastInsertId }
// dataHash is a Jabline Hash like { "name": "Alice", "role": "Admin" }
func dbInsert(args ...object.Object) object.Object {
	if len(args) != 3 {
		return newError("db_insert expects 3 arguments (conn, table, data)")
	}
	conn, ok := args[0].(*object.Database)
	if !ok {
		return newError("db_insert first argument must be a Database connection")
	}
	tableObj, ok := args[1].(*object.String)
	if !ok {
		return newError("db_insert second argument must be a STRING table name")
	}
	dataHash, ok := args[2].(*object.Hash)
	if !ok {
		return newError("db_insert third argument must be a HASH")
	}

	var cols []string
	var placeholders []string
	var vals []interface{}

	for _, pair := range dataHash.Pairs {
		if keyStr, ok := pair.Key.(*object.String); ok {
			cols = append(cols, keyStr.Value)
			placeholders = append(placeholders, "?")
			vals = append(vals, convertToNative(pair.Value))
		}
	}

	if len(cols) == 0 {
		return newError("db_insert: data hash is empty")
	}

	sql := "INSERT INTO " + tableObj.Value + " ("
	for i, c := range cols {
		if i > 0 { sql += ", " }
		sql += c
	}
	sql += ") VALUES ("
	for i, p := range placeholders {
		if i > 0 { sql += ", " }
		sql += p
	}
	sql += ")"

	res, err := conn.DB.Exec(sql, vals...)
	if err != nil {
		return newError("db_insert exec failed: %s", err.Error())
	}

	affected, _ := res.RowsAffected()
	lastId, _ := res.LastInsertId()

	hash := make(map[object.HashKey]object.HashPair)
	k1 := &object.String{Value: "rowsAffected"}
	hash[k1.HashKey()] = object.HashPair{Key: k1, Value: &object.Integer{Value: affected}}
	k2 := &object.String{Value: "lastInsertId"}
	hash[k2.HashKey()] = object.HashPair{Key: k2, Value: &object.Integer{Value: lastId}}

	return &object.Hash{Pairs: hash}
}
