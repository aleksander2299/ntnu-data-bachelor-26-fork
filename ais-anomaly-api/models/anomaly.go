package models

import (
	"encoding/json"
	"time"
)

// GeoJSON types
type GeoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}

type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	Geometry   GeoJSONGeometry        `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

type GeoJSONGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

// AnomalyGroup represents an anomaly group from the database
type AnomalyGroup struct {
	ID             int64     `json:"id"`
	Type           string    `json:"type"`
	MMSI           int64     `json:"mmsi"`
	StartedAt      time.Time `json:"startedAt"`
	LastActivityAt time.Time `json:"lastActivityAt"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
}

// ToGeoJSONFeature converts an AnomalyGroup to a GeoJSON Feature
func (ag *AnomalyGroup) ToGeoJSONFeature() GeoJSONFeature {
	return GeoJSONFeature{
		Type: "Feature",
		Geometry: GeoJSONGeometry{
			Type:        "Point",
			Coordinates: []float64{ag.Longitude, ag.Latitude},
		},
		Properties: map[string]interface{}{
			"id":             ag.ID,
			"type":           ag.Type,
			"mmsi":           ag.MMSI,
			"startedAt":      ag.StartedAt,
			"lastActivityAt": ag.LastActivityAt,
		},
	}
}

// AnomalyGroupsToGeoJSON converts a slice of AnomalyGroups to a GeoJSON FeatureCollection
func AnomalyGroupsToGeoJSON(groups []AnomalyGroup) GeoJSONFeatureCollection {
	features := make([]GeoJSONFeature, len(groups))
	for i, ag := range groups {
		features[i] = ag.ToGeoJSONFeature()
	}
	return GeoJSONFeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}
}

// Anomaly represents an individual anomaly from the database
type Anomaly struct {
	ID             int64           `json:"id"`
	Type           string          `json:"type"`
	Metadata       json.RawMessage `json:"metadata"`
	CreatedAt      time.Time       `json:"createdAt"`
	MMSI           *int64          `json:"mmsi,omitempty"`
	AnomalyGroupID *int64          `json:"anomalyGroupId,omitempty"`
	DataSource     string          `json:"dataSource"`
}

// AnomalyGroupWithAnomalies represents an anomaly group with its associated anomalies
type AnomalyGroupWithAnomalies struct {
	AnomalyGroup
	Anomalies []Anomaly `json:"anomalies"`
}

// AnomalyGroupsResponse represents the API response for anomaly groups
type AnomalyGroupsResponse struct {
	Data       []AnomalyGroup `json:"data"`
	TotalCount int            `json:"totalCount"`
	StartDate  string         `json:"startDate"`
	EndDate    string         `json:"endDate"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
