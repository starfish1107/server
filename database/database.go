package database

import (
	"os"
	"path/filepath"

	"github.com/gotify/server/auth/password"
	"github.com/gotify/server/model"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"    // enable the mysql dialect
	_ "github.com/jinzhu/gorm/dialects/postgres" // enable the postgres dialect
	_ "github.com/jinzhu/gorm/dialects/sqlite"   // enable the sqlite3 dialect
)

var mkdirAll = os.MkdirAll

// New creates a new wrapper for the gorm database framework.
func New(dialect, connection, defaultUser, defaultPass string, strength int, createDefaultUser bool) (*GormDatabase, error) {
	createDirectoryIfSqlite(dialect, connection)

	db, err := gorm.Open(dialect, connection)
	if err != nil {
		return nil, err
	}

	// We normally don't need that much connections, so we limit them. F.ex. mysql complains about
	// "too many connections", while load testing Gotify.
	db.DB().SetMaxOpenConns(10)

	if dialect == "sqlite3" {
		// We use the database connection inside the handlers from the http
		// framework, therefore concurrent access occurs. Sqlite cannot handle
		// concurrent writes, so we limit sqlite to one connection.
		// see https://github.com/mattn/go-sqlite3/issues/274
		db.DB().SetMaxOpenConns(1)
	}

	if !db.HasTable(new(model.User)) && !db.HasTable(new(model.Message)) &&
		!db.HasTable(new(model.Client)) && !db.HasTable(new(model.Application)) {
		db.AutoMigrate(new(model.User), new(model.Application), new(model.Message), new(model.Client))
		if createDefaultUser {
			db.Create(&model.User{Name: defaultUser, Pass: password.CreatePassword(defaultPass, strength), Admin: true})
		}
	}

	return &GormDatabase{DB: db}, nil
}

func createDirectoryIfSqlite(dialect string, connection string) {
	if dialect == "sqlite3" {
		if _, err := os.Stat(filepath.Dir(connection)); os.IsNotExist(err) {
			if err := mkdirAll(filepath.Dir(connection), 0777); err != nil {
				panic(err)
			}
		}
	}
}

// GormDatabase is a wrapper for the gorm framework.
type GormDatabase struct {
	DB *gorm.DB
}

// Close closes the gorm database connection.
func (d *GormDatabase) Close() {
	d.DB.Close()
}
