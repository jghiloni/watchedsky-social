package geojson_test

import (
	"encoding/json"
	"os"

	"github.com/jghiloni/watchedsky-social/backend/features"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Geometry", func() {
	Describe("Polygons", func() {
		It("Parses JSON", func() {
			f, e := os.Open("testdata/polygon.json")
			Expect(e).NotTo(HaveOccurred())
			defer f.Close()

			var p features.Feature
			e = json.NewDecoder(f).Decode(&p)
			Expect(e).NotTo(HaveOccurred())
		})
	})

	Describe("MultiPolygons", func() {
		It("Parses JSON", func() {
			f, e := os.Open("testdata/multipolygon.json")
			Expect(e).NotTo(HaveOccurred())
			defer f.Close()

			var p features.Feature
			e = json.NewDecoder(f).Decode(&p)
			Expect(e).NotTo(HaveOccurred())
		})
	})

	Describe("GeometryCollection", func() {
		It("Parses JSON", func() {
			f, e := os.Open("testdata/geometrycollection.json")
			Expect(e).NotTo(HaveOccurred())
			defer f.Close()

			var p features.Feature
			e = json.NewDecoder(f).Decode(&p)
			Expect(e).NotTo(HaveOccurred())
		})
	})
})
