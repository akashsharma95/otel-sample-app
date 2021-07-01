package db

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/razorpay/goutils/tracing/core"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"notesapp/internal/models"
	"notesapp/pkg/tracing"
)

const (
	contextKey = "ctx"
)

var connection *gorm.DB

func getDbConnection() *gorm.DB {
	if connection == nil {
		db, err := gorm.Open(sqlite.Open("notes.db"), &gorm.Config{})

		if err != nil {
			panic("failed to connect database")
		}

		err = db.AutoMigrate(&models.Note{})
		if err != nil {
			log.Fatalf("failed to run migrations %v", err)
		}
		RegisterCallbacks(db)

		connection = db
	}
	return connection
}

// GetDB - Database connection
func GetDB(ctx context.Context) *gorm.DB {
	return WithContext(ctx, getDbConnection())
}

// WithContext sets span to gorm settings, returns cloned DB
func WithContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	return db.Set(contextKey, ctx)
}

type callbacks struct{}

func newCallbacks() *callbacks {
	return &callbacks{}
}

func RegisterCallbacks(db *gorm.DB) {
	callbacks := newCallbacks()
	registerCallbacks(db, "create", callbacks)
	registerCallbacks(db, "query", callbacks)
	registerCallbacks(db, "update", callbacks)
	registerCallbacks(db, "delete", callbacks)
	registerCallbacks(db, "row_query", callbacks)
}

func (c *callbacks) beforeCreate(db *gorm.DB)   { c.before(db, "INSERT") }
func (c *callbacks) afterCreate(db *gorm.DB)    { c.after(db, "INSERT") }
func (c *callbacks) beforeQuery(db *gorm.DB)    { c.before(db, "SELECT") }
func (c *callbacks) afterQuery(db *gorm.DB)     { c.after(db, "SELECT") }
func (c *callbacks) beforeUpdate(db *gorm.DB)   { c.before(db, "UPDATE") }
func (c *callbacks) afterUpdate(db *gorm.DB)    { c.after(db, "UPDATE") }
func (c *callbacks) beforeDelete(db *gorm.DB)   { c.before(db, "DELETE") }
func (c *callbacks) afterDelete(db *gorm.DB)    { c.after(db, "DELETE") }
func (c *callbacks) beforeRowQuery(db *gorm.DB) { c.before(db, "") }
func (c *callbacks) afterRowQuery(db *gorm.DB)  { c.after(db, "") }

func (c *callbacks) before(db *gorm.DB, operation string) {
	ctx, ok := db.Get(contextKey)
	if !ok {
		return
	}

	spanName := fmt.Sprintf("%s - %s", operation, db.Statement.Table)
	_, newCtx := tracing.CreateChildSpan(ctx.(context.Context), spanName, otel.Tracer(""),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.KeyValue{
				Key:   semconv.DBSystemKey,
				Value: attribute.StringValue("sqlite"),
			},
			attribute.String(core.DbSystem, "sqlite"),
			attribute.String(core.SpanKind, core.Client),
			attribute.String(core.DbPeer, "notes.db"),
			attribute.String(core.DbPort, "3306"),
			attribute.String(core.DbConnection, "/notes.db"),
			attribute.String(core.DbName, "notes"),
		))

	db.Set(contextKey, newCtx)
}

func (c *callbacks) after(db *gorm.DB, operation string) {
	time.Sleep(time.Duration(rand.Int31n(5000)) * time.Millisecond)

	ctx, ok := db.Get(contextKey)
	if !ok {
		log.Println("failed to get contextKey")
		return
	}

	sql := db.Statement.SQL.String()

	if operation == "" && len(sql) > 0 {
		operation = strings.ToUpper(strings.Split(sql, " ")[0])
	}

	span := tracing.GetSpan(ctx.(context.Context))

	span.SetAttributes(attribute.String(core.DbStatement, sql))
	span.SetAttributes(attribute.String(core.DbTable, db.Statement.Table))
	span.SetAttributes(attribute.String(core.DbMethod, operation))
	span.SetAttributes(attribute.String(core.DbOperation, operation))
	span.SetAttributes(attribute.Any(core.DbCount, db.RowsAffected))

	if db.Error != nil && db.Error.Error() != core.RecordNotFound {
		tracing.RecordError(ctx.(context.Context), db.Error)
	}

	span.End()
}

func registerCallbacks(db *gorm.DB, name string, c *callbacks) {
	beforeName := fmt.Sprintf("tracing:%v_before", name)
	afterName := fmt.Sprintf("tracing:%v_after", name)
	gormCallbackName := fmt.Sprintf("gorm:%v", name)

	switch name {
	case "create":
		_ = db.Callback().Create().Before(gormCallbackName).Register(beforeName, c.beforeCreate)
		_ = db.Callback().Create().After(gormCallbackName).Register(afterName, c.afterCreate)
	case "query":
		_ = db.Callback().Query().Before(gormCallbackName).Register(beforeName, c.beforeQuery)
		_ = db.Callback().Query().After(gormCallbackName).Register(afterName, c.afterQuery)
	case "update":
		_ = db.Callback().Update().Before(gormCallbackName).Register(beforeName, c.beforeUpdate)
		_ = db.Callback().Update().After(gormCallbackName).Register(afterName, c.afterUpdate)
	case "delete":
		_ = db.Callback().Delete().Before(gormCallbackName).Register(beforeName, c.beforeDelete)
		_ = db.Callback().Delete().After(gormCallbackName).Register(afterName, c.afterDelete)
	case "row_query":
		_ = db.Callback().Row().Before(gormCallbackName).Register(beforeName, c.beforeRowQuery)
		_ = db.Callback().Row().After(gormCallbackName).Register(afterName, c.afterRowQuery)
	}
}
