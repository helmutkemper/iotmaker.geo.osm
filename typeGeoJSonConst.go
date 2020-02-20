package iotmaker_geo_osm

type GeoJSonType int

const (
	GEOJSON_POINT GeoJSonType = iota
	GEOJSON_LINE_STRING
	GEOJSON_POLYGON
	GEOJSON_MULTI_POINT
	GEOJSON_MULTI_LINE_STRING
	GEOJSON_MULTI_POLYGON
	GEOJSON_GEOMETRY_COLLECTION
)

var geoTypes = [...]string{
	"Point",
	"LineString",
	"Polygon",
	"MultiPoint",
	"MultiLineString",
	"MultiPolygon",
	"GeometryCollection",
}

func (e GeoJSonType) String() string {
	return geoTypes[e]
}
