// Package gorm 适配来自 https://github.com/glebarez/sqlite
package gorm

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten/core/fio"

	"gorm.io/gorm"
)

type (
	DB    struct{ *gorm.DB }
	Model struct{ *gorm.Model }
)

// New initialize a new sqlite3 db
func New(dsn string) gorm.Dialector {
	return new(Dialector{DSN: dsn})
}

// Open initialize a new db connection, need to import driver first
func Open[T string | fio.Path](dialect string, path T) (*DB, error) {
	if !strings.Contains(dialect, `sqlite`) {
		return nil, fmt.Errorf(`不支持的数据库类型喵！%w`, errors.ErrUnsupported)
	}
	db, err := gorm.Open(New(string(path)), new(gorm.Config))
	return new(DB{DB: db}), err
}

// AutoMigrate run auto migration for given models, will only add missing fields, won't delete/change current data
func (s *DB) AutoMigrate(values ...any) *DB {
	if err := s.DB.AutoMigrate(values...); err != nil {
		// 遇到可忽略错误时不 panic
		if !strings.Contains(err.Error(), `already exists`) {
			panic(err)
		}
		slog.Error(`数据库迁移`, slog.Any(`错误`, err))
	}
	return s
}

// Model specify the model you would like to run db operations
//
//	// update all users's name to `hello`
//	db.Model(&User{}).Update("name", "hello")
//	// if user's primary key is non-blank, will use it as condition, then will only update the user's name to `hello`
//	db.Model(&user).Update("name", "hello")
func (s *DB) Model(value any) *DB {
	return new(DB{DB: s.DB.Model(value)})
}

// Find find records that match given conditions
func (s *DB) Find(out any, where ...any) *DB {
	return new(DB{DB: s.DB.Find(out, where...)})
}

// First find first record that match given conditions, order by primary key
func (s *DB) First(out any, where ...any) *DB {
	return new(DB{DB: s.DB.First(out, where...)})
}

// Create insert the value into database
func (s *DB) Create(value any) *DB {
	return new(DB{DB: s.DB.Create(value)})
}

// Where return a new relation, filter records with given conditions, accepts `map`, `struct` or `string` as conditions, refer http://jinzhu.github.io/gorm/crud.html#query
func (s *DB) Where(query any, args ...any) *DB {
	return new(DB{DB: s.DB.Where(query, args...)})
}

// Update update attributes with callbacks, refer: https://jinzhu.github.io/gorm/crud.html#update
// WARNING when update with struct, GORM will not update fields that with zero value
func (s *DB) Update(attrs ...any) *DB {
	return new(DB{DB: s.DB.Updates(toSearchableMap(attrs...))})
}

func toSearchableMap(attrs ...any) (result any) {
	if len(attrs) > 1 {
		if str, ok := attrs[0].(string); ok {
			return map[string]any{str: attrs[1]}
		}
		return
	}
	if len(attrs) == 1 {
		if attr, ok := attrs[0].(map[string]any); ok {
			return attr
		}
		return attrs[0]
	}
	return
}

// Updates update attributes with callbacks, refer: https://jinzhu.github.io/gorm/crud.html#update
func (s *DB) Updates(values any, _ ...bool) *DB {
	return new(DB{DB: s.DB.Updates(values)})
}

// Close close current db connection.  If database connection is not an io.Closer, returns an error.
func (s *DB) Close() error {
	db, err := s.DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

// Table specify the table you would like to run db operations
func (s *DB) Table(name string) *DB {
	return new(DB{DB: s.DB.Table(name)})
}

// Take return a record that match given conditions, the order will depend on the database implementation
func (s *DB) Take(out any, where ...any) *DB {
	return new(DB{DB: s.DB.Take(out, where...)})
}

// Order specify order when retrieve records from database, set reorder to `true` to overwrite defined conditions
//
//	db.Order("name DESC")
//	db.Order("name DESC", true) // reorder
//	db.Order(gorm.Expr("name = ? DESC", "first")) // sql expression
func (s *DB) Order(value any, _ ...bool) *DB {
	return new(DB{DB: s.DB.Order(value)})
}

// Count get how many records for a model
func (s *DB) Count(value any) *DB {
	var (
		vn int64
		db = new(DB{DB: s.DB.Count(&vn)})
	)
	// 将计数结果赋值给传入的 value
	switch v := value.(type) {
	case *int:
		*v = int(vn)
	case *int8:
		*v = int8(vn)
	case *int16:
		*v = int16(vn)
	case *int32:
		*v = int32(vn)
	case *int64:
		*v = vn
	case *uint:
		*v = uint(vn)
	case *uint8:
		*v = uint8(vn)
	case *uint16:
		*v = uint16(vn)
	case *uint32:
		*v = uint32(vn)
	case *uint64:
		*v = uint64(vn)
	case *uintptr:
		*v = uintptr(vn)
	case *float32:
		*v = float32(vn)
	case *float64:
		*v = float64(vn)
	}
	return db
}
