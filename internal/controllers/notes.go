package controllers

import (
	"net/http"

	"notesapp/internal/db"
	"notesapp/internal/models"
	"notesapp/pkg/tracing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
)

// GetNote - get note by id
func GetNote(c *gin.Context) {
	span, ctx := tracing.CreateChildSpan(c.Request.Context(), "GetNote", otel.Tracer(""))
	if span != nil {
		m, _ := baggage.NewMember("key1", "value1")
		b, _ := baggage.New(m)
		baggage.ContextWithBaggage(ctx, b)
		defer span.End()
	}

	tracing.AddEvent(ctx, "FetchNote", []attribute.KeyValue{
		attribute.Any("testkey", "testvalue"),
	})

	var note models.Note
	id := c.Params.ByName("id")

	db := db.GetDB(ctx)

	if err := db.First(&note, id).Error; err != nil {
		c.JSON(
			http.StatusOK,
			gin.H{
				"error": "Note not found",
			},
		)
	} else {
		c.JSON(http.StatusOK, note)
	}
}

// GetNotes - get all notes
func GetNotes(c *gin.Context) {
	span, ctx := tracing.CreateChildSpan(c.Request.Context(), "GetNotes", otel.Tracer(""))
	if span != nil {
		defer span.End()
	}

	var notes []models.Note
	db := db.GetDB(ctx)

	if err := db.Find(&notes).Error; err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusOK, notes)
	}
}

// CreateNote - create note
func CreateNote(c *gin.Context) {
	span, ctx := tracing.CreateChildSpan(c.Request.Context(), "CreateNote", otel.Tracer(""))
	if span != nil {
		defer span.End()
	}

	var note models.Note
	db := db.GetDB(ctx)
	c.BindJSON(&note)

	if err := db.Create(&note).Error; err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusOK, note)
	}
}

// UpdateNote - update note
func UpdateNote(c *gin.Context) {
	span, ctx := tracing.CreateChildSpan(c.Request.Context(), "UpdateNote", otel.Tracer(""))
	if span != nil {
		defer span.End()
	}

	var note models.Note
	id := c.Params.ByName("id")

	db := db.GetDB(ctx)

	if err := db.First(&note, id).Error; err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	c.BindJSON(&note)

	if err := db.Save(&note).Error; err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusOK, note)
	}
}

// DeleteNote - delete note by id
func DeleteNote(c *gin.Context) {
	span, ctx := tracing.CreateChildSpan(c.Request.Context(), "DeleteNote", otel.Tracer(""))
	if span != nil {
		defer span.End()
	}

	var note models.Note
	id := c.Params.ByName("id")

	db := db.GetDB(ctx)

	if err := db.First(&note, id).Error; err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
	}

	if err := db.Delete(&note).Error; err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "Deleted",
		})
	}
}
