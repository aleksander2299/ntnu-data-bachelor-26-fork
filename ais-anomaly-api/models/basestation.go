package models

// BaseStation represents a base station along the Norwegian coast
type BaseStation struct {
	ID        int64   `json:"id" example:"1"`
	Name      string  `json:"name" example:"Stavanger"`
	Latitude  float64 `json:"latitude" example:"58.96463228151012"`
	Longitude float64 `json:"longitude" example:"5.738597994470069"`
}

// ToGeoJSONFeature converts a BaseStation to a GeoJSON Feature
func (bs *BaseStation) ToGeoJSONFeature() GeoJSONFeature {
	return GeoJSONFeature{
		Type: "Feature",
		Geometry: GeoJSONGeometry{
			Type:        "Point",
			Coordinates: []float64{bs.Longitude, bs.Latitude},
		},
		Properties: map[string]interface{}{
			"id":   bs.ID,
			"name": bs.Name,
		},
	}
}

// BaseStationsToGeoJSON converts a slice of BaseStations to a GeoJSON FeatureCollection
func BaseStationsToGeoJSON(stations []BaseStation) GeoJSONFeatureCollection {
	features := make([]GeoJSONFeature, len(stations))
	for i, bs := range stations {
		features[i] = bs.ToGeoJSONFeature()
	}
	return GeoJSONFeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}
}
