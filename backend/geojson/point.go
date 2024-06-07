package geojson

import (
	"encoding/json"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

type Point Coordinate

func (p Point) Type() GeometryType {
	return PointGeometryType
}

type intermediateGeometry struct {
	Type        string `json:"type" bson:"type"`
	Coordinates any    `json:"coordinates" bson:"coordinates"`
}

func defaultMarshal(g GeometryType, c any, marshaler coreMarshaler) ([]byte, error) {
	return marshaler(intermediateGeometry{
		Type:        string(g),
		Coordinates: c,
	})
}

func (p Point) MarshalJSON() ([]byte, error) {
	return defaultMarshal(p.Type(), Coordinate(p), json.Marshal)
}

func (p Point) MarshalBSON() ([]byte, error) {
	return defaultMarshal(p.Type(), Coordinate(p), bson.Marshal)
}

func (p *Point) unmarshal(data []byte, unmarshal coreUnmarshaler) error {
	var m map[string]any
	if err := unmarshal(data, &m); err != nil {
		return err
	}

	if m["type"] != string(PointGeometryType) {
		return fmt.Errorf("unexpected geometry type %q", m["type"])
	}

	l, ok := m["coordinates"].([]float64)
	if !ok {
		return errors.New("incorrect coordinate format")
	}

	*p = Point{Longitude: l[0], Latitude: l[1]}
	return nil
}

func (p *Point) UnmarshalJSON(data []byte) error {
	return p.unmarshal(data, json.Unmarshal)
}

func (p *Point) UnmarshalBSON(data []byte) error {
	return p.unmarshal(data, bson.Unmarshal)
}

type MultiPoint []Coordinate

func (m MultiPoint) Type() GeometryType {
	return MultiPointGeometryType
}

func (p MultiPoint) MarshalJSON() ([]byte, error) {
	return defaultMarshal(p.Type(), []Coordinate(p), json.Marshal)
}

func (p MultiPoint) MarshalBSON() ([]byte, error) {
	return defaultMarshal(p.Type(), []Coordinate(p), bson.Marshal)
}

func (p *MultiPoint) unmarshal(data []byte, unmarshal coreUnmarshaler) error {
	var m map[string]any
	if err := unmarshal(data, &m); err != nil {
		return err
	}

	if m["type"] != string(MultiPointGeometryType) {
		return fmt.Errorf("unexpected geometry type %q", m["type"])
	}

	coords, ok := m["coordinates"]
	if !ok {
		return errors.New("missing coordinates")
	}

	points, err := getLineOrMultipoint(coords)
	*p = points
	return err
}

func (p *MultiPoint) UnmarshalJSON(data []byte) error {
	return p.unmarshal(data, json.Unmarshal)
}

func (p *MultiPoint) UnmarshalBSON(data []byte) error {
	return p.unmarshal(data, bson.Unmarshal)
}
