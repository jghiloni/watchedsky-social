package geojson

import (
	"encoding/json"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

type Polygon [][]Coordinate

func (p Polygon) Type() GeometryType {
	return PolygonGeometryType
}

func (p Polygon) MarshalJSON() ([]byte, error) {
	return defaultMarshal(p.Type(), [][]Coordinate(p), json.Marshal)
}

func (p Polygon) MarshalBSON() ([]byte, error) {
	return defaultMarshal(p.Type(), [][]Coordinate(p), bson.Marshal)
}

func (p *Polygon) unmarshal(data []byte, unmarshal coreUnmarshaler) error {
	var m map[string]any
	if err := unmarshal(data, &m); err != nil {
		return err
	}

	if m["type"] != string(PolygonGeometryType) {
		return fmt.Errorf("unexpected geometry type %q", m["type"])
	}

	coords, ok := m["coordinates"]
	if !ok {
		return errors.New("missing coordinates")
	}

	polygon, err := getPolygonOrMultiline(coords)
	*p = polygon
	return err
}

func (p *Polygon) UnmarshalJSON(data []byte) error {
	return p.unmarshal(data, json.Unmarshal)
}

func (p *Polygon) UnmarshalBSON(data []byte) error {
	return p.unmarshal(data, bson.Unmarshal)
}

type MultiPolygon [][][]Coordinate

func (m MultiPolygon) Type() GeometryType {
	return MultiPolygonGeometryType
}

func (p MultiPolygon) MarshalJSON() ([]byte, error) {
	return defaultMarshal(p.Type(), [][][]Coordinate(p), json.Marshal)
}

func (p MultiPolygon) MarshalBSON() ([]byte, error) {
	return defaultMarshal(p.Type(), [][][]Coordinate(p), bson.Marshal)
}

func (m *MultiPolygon) unmarshal(data []byte, unmarshal coreUnmarshaler) error {
	var d map[string]any
	if err := unmarshal(data, &d); err != nil {
		return err
	}

	if d["type"] != string(MultiPolygonGeometryType) {
		return fmt.Errorf("unexpected geometry type %q", d["type"])
	}

	coords, ok := d["coordinates"]
	if !ok {
		return errors.New("missing coordinates")
	}

	polygons, err := getMultiPolygon(coords)
	*m = polygons
	return err
}

func (m *MultiPolygon) UnmarshalJSON(data []byte) error {
	return m.unmarshal(data, json.Unmarshal)
}

func (m *MultiPolygon) UnmarshalBSON(data []byte) error {
	return m.unmarshal(data, bson.Unmarshal)
}
