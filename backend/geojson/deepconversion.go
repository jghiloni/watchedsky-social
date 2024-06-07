package geojson

import (
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func normalizeToSlice(data any) ([]any, bool) {
	switch a := data.(type) {
	case []any:
		return a, true
	case primitive.A:
		return []any(a), true
	default:
		return nil, false
	}
}

func toFloat64(datum any) (float64, bool) {
	switch f := datum.(type) {
	case uint:
		return float64(f), true
	case uint8:
		return float64(f), true
	case uint16:
		return float64(f), true
	case uint32:
		return float64(f), true
	case uint64:
		return float64(f), true
	case int:
		return float64(f), true
	case int8:
		return float64(f), true
	case int16:
		return float64(f), true
	case int32:
		return float64(f), true
	case int64:
		return float64(f), true
	case float32:
		return float64(f), true
	case float64:
		return f, true
	default:
		return 0, false
	}
}

func getPoint(data any) (Coordinate, error) {
	if data == nil {
		return Coordinate{}, errors.New("data cannot be nil")
	}

	f, ok := normalizeToSlice(data)
	if !ok {
		return Coordinate{}, fmt.Errorf("expected [lon, lat], got %T", data)
	}

	lon, lonok := toFloat64(f[0])
	lat, latok := toFloat64(f[1])

	if !lonok {
		return Coordinate{}, fmt.Errorf("longitude must be a float64, was a %T", f[0])
	}

	if !latok {
		return Coordinate{}, fmt.Errorf("latitutde must be a float64, was a %T", f[1])
	}

	return Coordinate{Longitude: lon, Latitude: lat}, nil
}

func getLineOrMultipoint(data any) ([]Coordinate, error) {
	if data == nil {
		return nil, errors.New("data cannot be nil")
	}

	ps, ok := normalizeToSlice(data)
	if !ok {
		return nil, fmt.Errorf("expected [][]float64, got %T", data)
	}

	pts := make([]Coordinate, 0, len(ps))
	for _, pt := range ps {
		ft, err := getPoint(pt)
		if err != nil {
			return nil, err
		}
		pts = append(pts, ft)
	}

	return pts, nil
}

func getPolygonOrMultiline(data any) ([][]Coordinate, error) {
	if data == nil {
		return nil, errors.New("data cannot be nil")
	}

	ps, ok := normalizeToSlice(data)
	if !ok {
		return nil, fmt.Errorf("expected [][][]float64, got %T", data)
	}

	lines := make([][]Coordinate, 0, len(ps))
	for _, line := range ps {
		mp, err := getLineOrMultipoint(line)
		if err != nil {
			return nil, err
		}

		lines = append(lines, mp)
	}

	return lines, nil
}

func getMultiPolygon(data any) ([][][]Coordinate, error) {
	if data == nil {
		return nil, errors.New("data cannot be nil")
	}

	ps, ok := normalizeToSlice(data)
	if !ok {
		return nil, fmt.Errorf("expected [][][][]float64, got %T", data)
	}

	polys := make([][][]Coordinate, 0, len(ps))
	for _, poly := range ps {
		polygon, err := getPolygonOrMultiline(poly)
		if err != nil {
			return nil, err
		}

		polys = append(polys, polygon)
	}

	return polys, nil
}
