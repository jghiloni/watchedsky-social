package geojson_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGeojson(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Geojson Suite")
}
