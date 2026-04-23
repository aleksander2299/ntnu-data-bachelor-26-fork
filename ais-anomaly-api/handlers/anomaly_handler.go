package handlers

import (
	"database/sql"
	"fmt"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kyv-ekstern/ntnu-bachelor-26-ais-anomaly-api/models"
)

// AnomalyHandler handles anomaly-related HTTP requests
type AnomalyHandler struct {
	db *sql.DB
}

// NewAnomalyHandler creates a new AnomalyHandler
func NewAnomalyHandler(db *sql.DB) *AnomalyHandler {
	return &AnomalyHandler{db: db}
}

// GetAnomalyGroups godoc
//
//	@Summary	Get anomaly groups with optional filters
//	@Tags		anomaly-groups
//	@Produce	json
//	@Param		start_date	query		string	false	"Start date (YYYY-MM-DD)"
//	@Param		end_date	query		string	false	"End date (YYYY-MM-DD)"
//	@Param		mmsi		query		int		false	"MMSI filter"
//	@Param		type		query		string	false	"Anomaly type filter"
//	@Success	200			{object}	models.GeoJSONFeatureCollection
//	@Failure	400			{object}	models.ErrorResponse
//	@Failure	500			{object}	models.ErrorResponse
//	@Router		/anomaly-groups [get]
func (h *AnomalyHandler) GetAnomalyGroups(c *fiber.Ctx) error {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	mmsiStr := c.Query("mmsi")
	anomalyType := c.Query("type")

	// Needs to start with a where query to make it able to append later
	query := `
		SELECT 
			id,
			type,
			mmsi,
			started_at, 
			last_activity_at,
			ST_Y(position) as latitude,
			ST_X(position) as longitude
		FROM anomaly_groups
		WHERE 1=1
	`
	var args []interface{}
	paramIndex := 1

	if startDateStr != "" {
		startDate, err := parseDate(startDateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Error:   "invalid_start_date",
				Message: "Invalid start_date format.",
			})
		}
		// Got help from AI to be able to add multiple lines to the query without error messages
		query += fmt.Sprintf(" AND started_at >= $%d", paramIndex)
		args = append(args, startDate)
		paramIndex++
	}

	if endDateStr != "" {
		endDate, err := parseDate(endDateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Error:   "invalid_end_date",
				Message: "Invalid end_date format. Use YYYY-MM-DD or RFC3339 format.",
			})
		}
		query += fmt.Sprintf(" AND started_at <= $%d", paramIndex)
		args = append(args, endDate)
		paramIndex++
	}

	if mmsiStr != "" {
		mmsi, err := strconv.ParseInt(mmsiStr, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Error:   "invalid_mmsi",
				Message: "Invalid MMSI format.",
			})
		}
		query += fmt.Sprintf(" AND mmsi = $%d", paramIndex)
		args = append(args, mmsi)
		paramIndex++
	}

	if anomalyType != "" {
		query += fmt.Sprintf(" AND type = $%d", paramIndex)
		args = append(args, anomalyType)
		paramIndex++
	}

	query += " ORDER BY started_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "databaseError",
			Message: "Failed to query anomalies",
		})
	}
	defer rows.Close()

	var anomalyGroups []models.AnomalyGroup

	for rows.Next() {
		var ag models.AnomalyGroup
		err := rows.Scan(&ag.ID, &ag.Type, &ag.MMSI, &ag.StartedAt, &ag.LastActivityAt, &ag.Latitude, &ag.Longitude)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
				Error:   "scanError",
				Message: "Failed to parse anamaly group data.",
			})
		}
		anomalyGroups = append(anomalyGroups, ag)
	}
	if anomalyGroups == nil {
		anomalyGroups = []models.AnomalyGroup{}
	}

	return c.JSON(models.AnomalyGroupsToGeoJSON(anomalyGroups))
}

// GetAnomalyGroupsByMMSI godoc
//
//	@Summary	Get anomaly groups by MMSI
//	@Tags		anomaly-groups
//	@Param		mmsi	path		int	true	"MMSI"
//	@Success	200		{object}	models.GeoJSONFeatureCollection
//	@Failure	400		{object}	models.ErrorResponse
//	@Failure	500		{object}	models.ErrorResponse
//	@Router		/anomaly-groups/mmsi/{mmsi} [get]
func (h *AnomalyHandler) GetAnomalyGroupsByMMSI(c *fiber.Ctx) error {
	mmsiStr := c.Params("mmsi")
	mmsi, err := strconv.ParseInt(mmsiStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "invalidMmsi",
			Message: "Invalid anomaly group mmsi.",
		})
	}

	// Query anomaly group
	query := `
		SELECT 
			id, 
			type, 
			mmsi, 
			started_at, 
			last_activity_at,
			ST_Y(position) as latitude,
			ST_X(position) as longitude
		FROM anomaly_groups
		WHERE mmsi = $1
		ORDER BY started_at DESC
	`

	rows, err := h.db.Query(query, mmsi)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "databaseError",
			Message: "Failed to query anomaly groups.",
		})
	}
	defer rows.Close()

	// Holding All GeoJSON Features
	var features []models.GeoJSONFeature

	for rows.Next() {
		var ag models.AnomalyGroup

		err := rows.Scan(
			&ag.ID,
			&ag.Type,
			&ag.MMSI,
			&ag.StartedAt,
			&ag.LastActivityAt,
			&ag.Latitude,
			&ag.Longitude,
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
				Error:   "scanError",
				Message: "Failed to parse anomaly group data.",
			})
		}

		// Used some help from AI for this query in the loop
		anomalyQuery := `
			SELECT 
				id, 
				type, 
				metadata, 
				created_at, 
				mmsi, 
				anomaly_group_id, 
				data_source,
				source_id,
				signal_strength
			FROM anomalies
			WHERE anomaly_group_id = $1
			ORDER BY created_at DESC
		`

		anomalyRows, err := h.db.Query(anomalyQuery, ag.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
				Error:   "databaseError",
				Message: "Failed to query anomalies.",
			})
		}

		// Used AI to help create the anomalyRows Loop to make sure data was handled and parsed properly
		var anomalies []models.Anomaly
		for anomalyRows.Next() {
			var aDB models.AnomalyDB
			err := anomalyRows.Scan(
				&aDB.ID,
				&aDB.Type,
				&aDB.Metadata,
				&aDB.CreatedAt,
				&aDB.MMSI,
				&aDB.AnomalyGroupID,
				&aDB.DataSource,
				&aDB.SourceID,
				&aDB.SignalStrength,
			)
			if err != nil {
				anomalyRows.Close()
				return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
					Error:   "scanError",
					Message: "Failed to parse anomaly data.",
				})
			}
			anomalies = append(anomalies, aDB.ToAPIAnomaly())
		}
		anomalyRows.Close()

		if anomalies == nil {
			anomalies = []models.Anomaly{}
		}

		// Convert anomalies to a slice of maps for GeoJSON properties
		anomalyData := make([]map[string]interface{}, len(anomalies))
		for i, a := range anomalies {
			anomalyMap := map[string]interface{}{
				"id":             a.ID,
				"metadata":       a.Metadata,
				"createdAt":      a.CreatedAt,
				"anomalyGroupId": a.AnomalyGroupID,
				"dataSource":     a.DataSource,
			}
			if a.SourceID != nil {
				anomalyMap["sourceId"] = *a.SourceID
			}
			if a.SignalStrength != nil {
				anomalyMap["signalStrength"] = *a.SignalStrength
			}
			anomalyData[i] = anomalyMap
		}

		// Build GeoJSON Feature with anomalies included
		feature := models.GeoJSONFeature{
			Type: "Feature",
			Geometry: models.GeoJSONGeometry{
				Type:        "Point",
				Coordinates: []float64{ag.Longitude, ag.Latitude},
			},
			Properties: map[string]interface{}{
				"id":             ag.ID,
				"type":           ag.Type,
				"mmsi":           ag.MMSI,
				"startedAt":      ag.StartedAt,
				"lastActivityAt": ag.LastActivityAt,
				"anomalies":      anomalyData,
			},
		}
		features = append(features, feature)
	}

	return c.JSON(models.GeoJSONFeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	})
}

// GetAnomalyGroupByID godoc
// @Summary Get anomaly group by ID
// @Tags anomaly-groups
// @Param id path int true "Anomaly Group ID"
// @Success 200 {object} models.GeoJSONFeatureCollection
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /anomaly-groups/{id} [get]
func (h *AnomalyHandler) GetAnomalyGroupByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "invalidId",
			Message: "Invalid anomaly group ID.",
		})
	}

	// Query anomaly group
	query := `
		SELECT 
			id, 
			type, 
			mmsi, 
			started_at, 
			last_activity_at,
			ST_Y(position) as latitude,
			ST_X(position) as longitude
		FROM anomaly_groups
		WHERE id = $1
	`

	var ag models.AnomalyGroup
	err = h.db.QueryRow(query, id).Scan(
		&ag.ID,
		&ag.Type,
		&ag.MMSI,
		&ag.StartedAt,
		&ag.LastActivityAt,
		&ag.Latitude,
		&ag.Longitude,
	)

	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error:   "notFound",
			Message: "Anomaly group not found.",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "databaseError",
			Message: "Failed to query anomaly group.",
		})
	}

	// Query anomalies for this group
	anomalyQuery := `
		SELECT 
			id, 
			type, 
			metadata, 
			created_at, 
			mmsi, 
			anomaly_group_id, 
			data_source,
			source_id,
			signal_strength
		FROM anomalies
		WHERE anomaly_group_id = $1
		ORDER BY created_at DESC
	`

	rows, err := h.db.Query(anomalyQuery, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "databaseError",
			Message: "Failed to query anomalies.",
		})
	}
	defer rows.Close()

	var anomalies []models.Anomaly
	for rows.Next() {
		var aDB models.AnomalyDB
		err := rows.Scan(
			&aDB.ID,
			&aDB.Type,
			&aDB.Metadata,
			&aDB.CreatedAt,
			&aDB.MMSI,
			&aDB.AnomalyGroupID,
			&aDB.DataSource,
			&aDB.SourceID,
			&aDB.SignalStrength,
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
				Error:   "scanError",
				Message: "Failed to parse anomaly data.",
			})
		}
		// Transform to API model (removing type and mmsi)
		anomalies = append(anomalies, aDB.ToAPIAnomaly())
	}

	// Build a FeatureCollection with the group's metadata at the top level
	// and one homogeneous feature per anomaly.
	features := make([]models.GeoJSONFeature, 0, len(anomalies))

	for _, a := range anomalies {
		// Use the first position report from metadata as the anomaly's geometry,
		// falling back to the group's position if none are present.
		lon, lat := ag.Longitude, ag.Latitude
		var meta struct {
			PositionReports []struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"positionReports"`
		}
		if err := json.Unmarshal(a.Metadata, &meta); err == nil && len(meta.PositionReports) > 0 {
			lon = meta.PositionReports[0].Longitude
			lat = meta.PositionReports[0].Latitude
		}

		props := map[string]interface{}{
			"id":             a.ID,
			"metadata":       a.Metadata,
			"createdAt":      a.CreatedAt,
			"anomalyGroupId": a.AnomalyGroupID,
			"dataSource":     a.DataSource,
		}
		if a.SourceID != nil {
			props["sourceId"] = *a.SourceID
		}
		if a.SignalStrength != nil {
			props["signalStrength"] = *a.SignalStrength
		}

		features = append(features, models.GeoJSONFeature{
			Type: "Feature",
			Geometry: models.GeoJSONGeometry{
				Type:        "Point",
				Coordinates: []float64{lon, lat},
			},
			Properties: props,
		})
	}

	return c.JSON(models.GeoJSONFeatureCollection{
		Type: "FeatureCollection",
		Properties: map[string]interface{}{
			"id":             ag.ID,
			"type":           ag.Type,
			"mmsi":           ag.MMSI,
			"startedAt":      ag.StartedAt,
			"lastActivityAt": ag.LastActivityAt,
		},
		Features: features,
	})
}

// parseDate attempts to parse a date string in multiple formats
func parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fiber.NewError(fiber.StatusBadRequest, "invalid date format")
}
