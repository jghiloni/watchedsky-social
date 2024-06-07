package geojson

import (
	"encoding/json"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

type LineString []Coordinate

func (l LineString) Type() GeometryType {
	return LineStringGeometryType
}

func (p LineString) MarshalJSON() ([]byte, error) {
	return defaultMarshal(p.Type(), []Coordinate(p), json.Marshal)
}

func (p LineString) MarshalBSON() ([]byte, error) {
	return defaultMarshal(p.Type(), []Coordinate(p), bson.Marshal)
}

func (l *LineString) unmarshal(data []byte, unmarshal coreUnmarshaler) error {
	var m map[string]any
	if err := unmarshal(data, &m); err != nil {
		return err
	}

	if m["type"] != string(LineStringGeometryType) {
		return fmt.Errorf("unexpected geometry type %q", m["type"])
	}

	coords, ok := m["coordinates"]
	if !ok {
		return errors.New("missing coordinates")
	}

	line, err := getLineOrMultipoint(coords)
	*l = line
	return err
}

func (l *LineString) UnmarshalJSON(data []byte) error {
	return l.unmarshal(data, json.Unmarshal)
}

func (l *LineString) UnmarshalBSON(data []byte) error {
	return l.unmarshal(data, bson.Unmarshal)
}

type MultiLineString [][]Coordinate

func (m MultiLineString) Type() GeometryType {
	return MultiLineStringGeometryType
}

func (m MultiLineString) MarshalJSON() ([]byte, error) {
	return defaultMarshal(m.Type(), [][]Coordinate(m), json.Marshal)
}

func (m MultiLineString) MarshalBSON() ([]byte, error) {
	return defaultMarshal(m.Type(), [][]Coordinate(m), bson.Marshal)
}

func (m *MultiLineString) unmarshal(data []byte, unmarshal coreUnmarshaler) error {
	var d map[string]any
	if err := unmarshal(data, &d); err != nil {
		return err
	}

	if d["type"] != string(MultiLineStringGeometryType) {
		return fmt.Errorf("unexpected geometry type %q", d["type"])
	}

	coords, ok := d["coordinates"]
	if !ok {
		return errors.New("missing coordinates")
	}

	lines, err := getPolygonOrMultiline(coords)
	*m = lines
	return err
}

func (m *MultiLineString) UnmarshalJSON(data []byte) error {
	return m.unmarshal(data, json.Unmarshal)
}

func (m *MultiLineString) UnmarshalBSON(data []byte) error {
	return m.unmarshal(data, bson.Unmarshal)
}
