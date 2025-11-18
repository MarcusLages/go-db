# GO + PostgreSQL (but for general DB)
https://go.dev/doc/tutorial/database-access
https://github.com/jackc/pgx/wiki/Getting-started-with-pgx-through-database-sql
### Concepts
#### `database/sql` and SQL Drivers
- Abstraction layer to define common functions to handle SQL database handling
- To actually implement the functions in `database/sql`, you need to choose a GO driver
	- You can find most of them in the [GO Wiki: SQL Database Drivers page](https://go.dev/wiki/SQLDrivers)
- For our case (PostgreSQL), we will use the [pgx](https://github.com/jackc/pgx) package by [Jack Christensen](https://github.com/jackc)
- For all the next examples, we will be using PostgreSQL, but for the GO steps, the only thing you have to do differently is choosing and importing your SQL driver for your specific DB
#### Connection Strings
- Strings that have the necessary data to connect to your database.
- The necessary data is usually in a `.env` file and it's usually/should be reconstructed by the database drivers/libraries
- Needed info:
	- Database type used (PostgreSQL, MySQL, etc.)
	- User
	- Password
	- Host/IP address
	- Port number (different default ones for different DBs, for PostgreSQL, the default port is 5432)
	- Database name
- I usually use the go library [godotenv](https://github.com/joho/godotenv) to load my environmental variables
```bash
db_type://user:password@host:port/database_name
```
#### Query, QueryRow, Exec and Prepare
- Core concept of `databaase/sql` since all DBs should implement it
- `Query()` is for multi-row querying
- `QueryRow()` is for single-row querying
- `Exec()` is for executing a deliberate/free query
- `Prepare()` is for preparing queries before end
### Preparations
1. Start your DB service. Most of the times it's automatic by your OS or hosting provider, but it's always good to make sure.
2. Create your database.
	- Make sure you have a valid user
```postgresql
CREATE ROLE sql_user WITH LOGIN PASSWORD '1234';
CREATE DATABASE gen_db OWNER sql_user;
```
3. To testify everything is okay, you can try to connect to the database through shell connection or in your app through a connection string
```bash
psql -U sql_user -d gen_db

postgres://sql_user:password@localhost:5432/gen_db
```
### Set Up Steps
1. Choose your [SQL driver](https://go.dev/wiki/SQLDrivers)
	- We are choosing [pgx](https://github.com/jackc/pgx) version 5
2. Add it to your module using `go get <driver-library>`. 
	- For pgx, download `pgx` and `pgxpool` for integration with `database/sql` and download dependencies using `go mod tidy`
```bash
go mod tidy
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/stdlib
go get github.com/joho/godotenv
```
3. Import the package in your go module:
```go
import "github.com/jackc/pgx/v5"
```
### Connect to your DB
Different databases have different ways of connecting to them, but they all tend to use a connection string like above. 
1. Build the connection string 
	- You should always let your SQL driver handle that, but you can also create it yourself using the template above if your driver doesn't support that (like pgx). Here's some MySQL code from [Go tutorials](https://go.dev/doc/tutorial/database-access) for that:
```go
cfg := mysql.NewConfig()
    cfg.User = os.Getenv("DBUSER")
    cfg.Passwd = os.Getenv("DBPASS")
    cfg.Net = "tcp"
    cfg.Addr = "127.0.0.1:3306"
    cfg.DBName = "recordings"

// Create the connection URL
conn_url := cfg.FormatDSN()
log.Println("Connection url:", conn_url)

// OR:
func conn_url(user, passwd, host string, port int, db string) string {
    return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, passwd, host, port, db)
}
```
2. Use the `sql.Open()` function from `database/sql` to connect with the database by passing the SQL driver and the connection URL to it
```go
// Connect using the connection URL
    db, err := sql.Open("pgx", conn_url)
    if err != nil {
        log.Fatal(err)
    }
```
3. You may ping the db to make sure the connection is good
	- Also don't forget to close the connection. If you have one function that handles ALL the DB functionality, you can use defer to not forget about it
```go
// Defer the connection closing to when the function closes
// Common in functions that represent the whole execution of the program
defer db.Close()

err := db.Ping()
if err != nil {
	log.Println("Connected to database.")
}
```
### Creating a table using `Exec()`
- `database/sql` is just a basic interface, so it only has three functions `Exec()` to execute SQL statements, `Query/QueryRow()` to select data and `Prepare()` to prepare data
- Let's use `Exec()` and pass it a table creation SQL string and make it execute it
```go
_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS albums (
		id BIGSERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		artist TEXT NOT NULL,
		score REAL NOT NULL CHECK (score >= 0 AND score <= 10)
	)
`)
if err != nil {
	log.Fatal(err)
}
```
### Adding data with `Exec()` and `Prepare()`
- Very similar to before, but this time we will use `Prepare()` to check for early schema errors
```go
prep_q, err := db.Prepare(`
        INSERT INTO albums (title, artist, score) VALUES ($1, $2, $3)
    `)
    if err != nil {
        return fmt.Errorf("insert_data: %v", err)
    }
  
    _, err = prep_q.Exec(album.Title, album.Artist, album.Score)
    if err != nil {
        return fmt.Errorf("insert_data: %v", err)
    }
```
### Querying One Row
- For one row, we use a combination of `QueryRow()` to prepare the query for execution and `Scan()` for running it and getting the first value
- `Scan()` receives pointers to each data part of the return structure
- The return value of the `Scan()` is actually the error
```go
// Prepares the select query
row := db.QueryRow("SELECT * FROM album WHERE title = $1", title)

// Runs the query and reads only the first row with row.Scan
if err := row.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Score); err != nil {
	// Returns sql.ErrNoRows error if not found
	if err == sql.ErrNoRows {
		return alb, fmt.Errorf("no album with title %s", title)
	}
	return alb, fmt.Errorf("album_by_title %s: %v", title, err)
}
```
### Querying for Multiple Rows
- `Query()` is used to return a row lazy iterator 
- As a resource manager, you have to close it
- Still use `Scan()` in the same way
- Use `Err()` to check for errors
```go
var albums []Album

// Returns a query lazy iterator
rows, err := db.Query("SELECT * FROM albums WHERE artist = $1", name)
if err != nil {
	return nil, fmt.Errorf("album_by_artist %q: %v", name, err)
}
defer rows.Close()

for rows.Next() {
	var alb Album
	if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Score); err != nil {
		return nil, fmt.Errorf("albums_by_artist %q: %v", name, err)
	}
	albums = append(albums, alb)
}

// Returns non-nil if any error was found in the iterating process
if err := rows.Err(); err != nil {
	return albums, fmt.Errorf("albums_by_artist %q: %v", name, err)
}
return albums, nil
```
### Final Notes
- `database/sql` is pretty limited to very simple SQL queries
	- Maybe too much
- While its objective is to keep everything centralized for different databases, and it does achieve that, the fact that there's some features that are database dependant (like query parameters), make it a bit disappointing
- If you are dead set on a fix database, probably use that database's driver and library, but for simple tasks and flexible SQL queries `database/sql` is a good one