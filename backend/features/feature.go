package features

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jghiloni/watchedsky-social/backend/geojson"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	Alert string = "wx:Alert"
	Zone  string = "wx:Zone"

	CollectionName string = "features"
)

type JSONObject map[string]any

func (j JSONObject) StringValue(key string) string {
	s, _ := j[key].(string)
	return s
}

func (j JSONObject) IntValue(key string) int64 {
	i, _ := j[key].(int64)
	return i
}

func (j JSONObject) FloatValue(key string) float64 {
	f, _ := j[key].(float64)
	return f
}

type Feature struct {
	ID         string           `json:"id" bson:"_id"`
	Geometry   geojson.Geometry `json:"geometry"`
	Properties JSONObject       `json:"properties"`
}

type Features []Feature

func (f Features) Len() int {
	return len(f)
}

func (f Features) Less(i, j int) bool {
	// sort by sent time descending
	t1s, ok1 := f[i].Properties["sent"].(string)
	t2s, ok2 := f[j].Properties["sent"].(string)

	var t1, t2 = time.UnixMicro(0), time.UnixMicro(0)
	var err error
	if ok1 {
		t1, err = time.Parse(time.RFC3339, t1s)
		if err != nil {
			return false
		}
	}

	if ok2 {
		t2, err = time.Parse(time.RFC3339, t2s)
		if err != nil {
			return false
		}
	}

	return t1.After(t2)
}

func (f Features) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

type FeatureCollection struct {
	Features Features `json:"features"`
}

type intermediateFeature struct {
	ID         string         `json:"id" bson:"_id"`
	Geometry   map[string]any `json:"geometry"`
	Properties map[string]any `json:"properties"`
}

type unmarshaler func([]byte, any) error
type marshaler func(any) ([]byte, error)

func (f *Feature) unmarshal(data []byte, fmtMarshal marshaler, fmtUnmarshal unmarshaler) error {
	var intF intermediateFeature
	if err := fmtUnmarshal(data, &intF); err != nil {
		return err
	}

	(*f).ID = intF.ID
	(*f).Properties = intF.Properties

	if intF.Geometry != nil {
		if rawType, ok := intF.Geometry["type"]; ok {
			if geoType, ok := rawType.(string); ok {
				geoBytes, err := fmtMarshal(intF.Geometry)
				if err != nil {
					return err
				}

				switch geojson.GeometryType(geoType) {
				case geojson.PointGeometryType:
					g := geojson.Point{}
					if err = fmtUnmarshal(geoBytes, &g); err != nil {
						return err
					}

					(*f).Geometry = g
				case geojson.MultiPointGeometryType:
					g := geojson.MultiPoint{}
					if err = fmtUnmarshal(geoBytes, &g); err != nil {
						return err
					}

					(*f).Geometry = g
				case geojson.LineStringGeometryType:
					g := geojson.LineString{}
					if err = fmtUnmarshal(geoBytes, &g); err != nil {
						return err
					}

					(*f).Geometry = g
				case geojson.MultiLineStringGeometryType:
					g := geojson.MultiLineString{}
					if err = fmtUnmarshal(geoBytes, &g); err != nil {
						return err
					}

					(*f).Geometry = g
				case geojson.PolygonGeometryType:
					g := geojson.Polygon{}
					if err = fmtUnmarshal(geoBytes, &g); err != nil {
						return err
					}

					(*f).Geometry = g
				case geojson.MultiPolygonGeometryType:
					g := geojson.MultiPolygon{}
					if err = fmtUnmarshal(geoBytes, &g); err != nil {
						return err
					}

					(*f).Geometry = g
				case geojson.GeometryCollectionType:
					g := geojson.GeometryCollection{}
					if err = fmtUnmarshal(geoBytes, &g); err != nil {
						return err
					}

					(*f).Geometry = g
				default:
					return fmt.Errorf("unrecognized geometry type %q", geoType)
				}

			}
		}
	}

	return nil
}

func (f *Feature) UnmarshalJSON(data []byte) error {
	return f.unmarshal(data, json.Marshal, json.Unmarshal)
}

func (f *Feature) UnmarshalBSON(data []byte) error {
	return f.unmarshal(data, bson.Marshal, bson.Unmarshal)
}
