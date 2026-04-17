package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	srvConfig "github.com/CHESSComputing/golib/config"
	"github.com/gin-gonic/gin"
)

// SearchHandler handles GET /search?did=<value>[&text=<phrase>]
//
// Query logic:
//   - `did`  (required): exact match on the primary DID field.
//   - `text` (optional): MongoDB $text search across entries.text.
//
// Results are rendered via the "search.html" template with entries sorted
// newest-first inside the template layer (Go sort on the slice).
func SearchHandler(c *gin.Context) {
	var records []map[string]any
	did := c.Query("did")
	if did == "" {
		c.JSON(http.StatusBadRequest, records)
		return
	}
	text := c.Query("text")

	spec := map[string]any{"did": did}
	if text != "" {
		spec["$text"] = map[string]any{"$search": text}
	}
	skeys := []string{"date"}
	sOrder := -1 // 1 ascending, -1 descending
	records = metaDB.GetSorted(
		srvConfig.Config.ELogData.DBName,
		srvConfig.Config.ELogData.DBColl,
		spec, skeys, sOrder, 0, -1)
	if Verbose > 0 {
		log.Println("RecordHandler", spec, records)
	}
	c.JSON(http.StatusOK, records)
}

// UpdateHandler handles /update end-point
func UpdateHandler(c *gin.Context) {
	var did, user, text, image_url string

	// Detect content type
	contentType := c.GetHeader("Content-Type")

	if strings.HasPrefix(contentType, "application/json") {
		// JSON input
		var payload struct {
			DID      string `json:"did"`
			User     string `json:"user"`
			Text     string `json:"text"`
			ImageURL string `json:"image_url"`
		}

		if err := c.ShouldBindJSON(&payload); err != nil {
			log.Printf("ERROR: unable to bind JSON, error %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}

		did = payload.DID
		user = payload.User
		text = payload.Text
		image_url = payload.ImageURL

	} else {
		// Form input (default)
		did = c.PostForm("did")
		user = c.PostForm("user")
		text = c.PostForm("text")
		image_url = c.PostForm("image_url")
	}

	// Validate input
	if user == "" || did == "" {
		log.Printf("ERROR: either user=%s, did=%s are empty.", user, text, did)
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
		return
	}

	// Create record
	rec := map[string]any{
		"date":      time.Now().UnixNano(),
		"text":      text,
		"user":      user,
		"image_url": image_url,
		"did":       did,
	}

	metaDB.InsertRecord(
		srvConfig.Config.ELogData.DBName,
		srvConfig.Config.ELogData.DBColl,
		rec,
	)

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
