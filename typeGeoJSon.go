package iotmaker_geo_osm

import (
	"encoding/json"
	"math"
)

type point [3]float64

type lineString [3]float64

type polygon [][3]float64

type multiPoint [3]float64

type multiLineString [][3]float64

type multiPolygon [][][3]float64

type geometry struct {
	Type        string      `bson:"type" json:"type"`
	typeConst   GeoJSonType `bson:"-" json:"-"`
	BoundingBox interface{} `bson:"bbox,omitempty" json:"bbox,omitempty"`
	Coordinates interface{} `bson:"coordinates" json:"coordinates"`
}

type features struct {
	setOfCoordinates int               `bson:"-" json:"-"`
	setOfLines       int               `bson:"-" json:"-"`
	setOfPolygons    int               `bson:"-" json:"-"`
	Type             string            `bson:"type" json:"type"`
	Id               string            `bson:"id" json:"id"`
	Properties       map[string]string `bson:"properties" json:"properties"`
	Geometry         geometry          `bson:"geometry" json:"geometry"`
}

type featureJSon struct {
	Id   int64             `bson:"id" json:"-"`
	JSon string            `bson:"geoJSon" json:"-"`
	Tag  map[string]string `bson:"tag" json:"-"`
}

type GeoJSon struct {
	// id do open street maps
	Id            int64             `bson:"id" json:"-"`
	IdRelation    int64             `bson:"idRelation,omitempty" json:"-"`
	IdSubRelation int64             `bson:"idSubRelation,omitempty" json:"-"`
	JSon          string            `bson:"geoJSon" json:"-"`
	setOfFeatures int               `bson:"-" json:"-"`
	Type          string            `bson:"-" json:"type"`
	Features      []features        `bson:"-" json:"features"`
	Tag           map[string]string `bson:"tag" json:"-"`
}

func (e *GeoJSon) Init() {
	e.setOfFeatures = -1
	e.Type = "FeatureCollection"
	e.Features = make([]features, 0)
	e.Tag = make(map[string]string)
}

func (e *GeoJSon) AddTag(key, value string) {
	e.Tag[key] = value
}

func (e *GeoJSon) AddGeoMathWay(id string, way *WayStt) {
	e.NewFeature(id, GEOJSON_LINE_STRING)
	for _, coordinates := range way.Loc {
		e.AddLngLat(coordinates[0], coordinates[1])
	}
	for tagKey, tagValue := range way.Tag {
		e.AddProperties(tagKey, tagValue)
		e.AddTag(tagKey, tagValue)
	}
	e.MakeBoundingBox()
}

func (e *GeoJSon) AddGeoMathPoint(id string, point *PointStt) {
	e.NewFeature(id, GEOJSON_POINT)
	for tagKey, tagValue := range point.Tag {
		e.AddProperties(tagKey, tagValue)
		e.AddTag(tagKey, tagValue)
	}
	e.AddLngLat(point.Loc[0], point.Loc[1])
}

func (e *GeoJSon) AddGeoMathPolygon(id string, polygon *PolygonStt) {
	e.NewFeature(id, GEOJSON_POLYGON)
	for _, point := range polygon.PointsList {
		e.AddLngLat(point.Loc[0], point.Loc[1])
	}
	for tagKey, tagValue := range polygon.Tag {
		e.AddProperties(tagKey, tagValue)
		e.AddTag(tagKey, tagValue)
	}
	e.ClosePolygon()
	e.MakeBoundingBox()
}

func (e *GeoJSon) AddGeoMathPolygonList(id string, polygon *PolygonListStt) {
	e.NewFeature(id, GEOJSON_MULTI_POLYGON)
	e.SetOfMultiPolygons(len(polygon.List))
	for k, listOfPolygons := range polygon.List {

		if k != 0 {
			e.NewPolygon()
		}

		for _, point := range listOfPolygons.PointsList {
			e.AddLngLat(point.Loc[0], point.Loc[1])
		}
		for tagKey, tagValue := range listOfPolygons.Tag {
			e.AddProperties(tagKey, tagValue)
			e.AddTag(tagKey, tagValue)
		}
		e.ClosePolygon()
		e.MakeBoundingBox()
	}
}

func (e *GeoJSon) NewFeature(id string, geoType GeoJSonType) {
	if len(e.Features) == 0 {
		e.Features = make([]features, 0)
	}

	e.setOfFeatures += 1

	var f features = features{
		setOfCoordinates: 0,
		setOfLines:       0,
		setOfPolygons:    1,
		Id:               id,
		Type:             "Feature",
		Geometry: geometry{
			typeConst: geoType,
			Type:      geoType.String(),
		},
	}

	switch geoType {
	case GEOJSON_POINT:
		f.Geometry.Coordinates = []point{}
	case GEOJSON_LINE_STRING:
		f.Geometry.Coordinates = []lineString{}
	case GEOJSON_POLYGON:
		f.Geometry.Coordinates = []polygon{}
	case GEOJSON_MULTI_POINT:
		f.Geometry.Coordinates = []multiPoint{}
	case GEOJSON_MULTI_LINE_STRING:
		f.Geometry.Coordinates = []multiLineString{}
	case GEOJSON_MULTI_POLYGON:
		f.Geometry.Coordinates = []multiPolygon{}
	}

	e.Features = append(e.Features, f)
}

func (e *GeoJSon) AddProperties(key, value string) {
	if len(e.Features[e.setOfFeatures].Properties) == 0 {
		e.Features[e.setOfFeatures].Properties = make(map[string]string)
	}
	e.Features[e.setOfFeatures].Properties[key] = value
}

func (e *GeoJSon) NewPolygon() {
	e.Features[e.setOfFeatures].setOfLines += 1
}

func (e *GeoJSon) SetOfMultiPolygons(value int) {
	e.Features[e.setOfFeatures].setOfPolygons = value
}

func (e *GeoJSon) NewSetOfCoordinates() {
	e.Features[e.setOfFeatures].setOfCoordinates += 1
	e.Features[e.setOfFeatures].setOfLines = 0

	switch e.Features[e.setOfFeatures].Geometry.typeConst {
	case GEOJSON_POINT:
		e.Features[e.setOfFeatures].Geometry.Coordinates = append(e.Features[e.setOfFeatures].Geometry.Coordinates.([]point), point{})

	case GEOJSON_LINE_STRING:
		e.Features[e.setOfFeatures].Geometry.Coordinates = append(e.Features[e.setOfFeatures].Geometry.Coordinates.([]lineString), lineString{})

	case GEOJSON_POLYGON:
		e.Features[e.setOfFeatures].Geometry.Coordinates = append(e.Features[e.setOfFeatures].Geometry.Coordinates.([]polygon), polygon{})

	case GEOJSON_MULTI_POINT:
		e.Features[e.setOfFeatures].Geometry.Coordinates = append(e.Features[e.setOfFeatures].Geometry.Coordinates.([]multiPoint), multiPoint{})

	case GEOJSON_MULTI_LINE_STRING:
		e.Features[e.setOfFeatures].Geometry.Coordinates = append(e.Features[e.setOfFeatures].Geometry.Coordinates.([]multiLineString), multiLineString{})

	case GEOJSON_MULTI_POLYGON:
		e.Features[e.setOfFeatures].Geometry.Coordinates = append(e.Features[e.setOfFeatures].Geometry.Coordinates.([]multiPolygon), multiPolygon{})
	}
}

func (e *GeoJSon) ClosePolygon() {
	switch e.Features[e.setOfFeatures].Geometry.typeConst {
	case GEOJSON_POLYGON:
		firstPointPolygon := e.Features[e.setOfFeatures].Geometry.Coordinates.([]polygon)[0]
		e.Features[e.setOfFeatures].Geometry.Coordinates.([]polygon)[e.Features[e.setOfFeatures].setOfCoordinates] = append(e.Features[e.setOfFeatures].Geometry.Coordinates.([]polygon)[e.Features[e.setOfFeatures].setOfCoordinates], firstPointPolygon[0])

	case GEOJSON_MULTI_POLYGON:
		firstPointMultiPolygon := e.Features[e.setOfFeatures].Geometry.Coordinates.([]multiPolygon)[e.Features[e.setOfFeatures].setOfCoordinates][e.Features[e.setOfFeatures].setOfLines][0]
		e.Features[e.setOfFeatures].Geometry.Coordinates.([]multiPolygon)[e.Features[e.setOfFeatures].setOfCoordinates][e.Features[e.setOfFeatures].setOfLines] = append(e.Features[e.setOfFeatures].Geometry.Coordinates.([]multiPolygon)[e.Features[e.setOfFeatures].setOfCoordinates][e.Features[e.setOfFeatures].setOfLines], firstPointMultiPolygon)
	}
}

/*
  todo: The four lines of the bounding box are defined fully within the
  coordinate reference system; that is, for a box bounded by the values
  "west", "south", "east", and "north", every point on the northernmost
  line can be expressed as

  (lon, lat) = (west + (east - west) * t, north)

  with 0 <= t <= 1.
*/
func (e *GeoJSon) MakeBoundingBox() {
	var latMin, latMax, lngMin, lngMax float64

	switch e.Features[e.setOfFeatures].Geometry.typeConst {
	case GEOJSON_LINE_STRING:
		for k, v := range e.Features[e.setOfFeatures].Geometry.Coordinates.([]lineString) {
			if k == 0 {
				latMin = v[1]
				latMax = v[1]

				lngMin = v[0]
				lngMax = v[0]
			} else {
				latMin = math.Min(latMin, v[1])
				latMax = math.Max(latMax, v[1])

				lngMin = math.Min(lngMin, v[0])
				lngMax = math.Max(lngMax, v[0])
			}
		}

	case GEOJSON_POLYGON:
		for k, v := range e.Features[e.setOfFeatures].Geometry.Coordinates.([]polygon) {
			if k == 0 {
				latMin = v[k][1]
				latMax = v[k][1]

				lngMin = v[k][0]
				lngMax = v[k][0]
			} else {
				latMin = math.Min(latMin, v[k][1])
				latMax = math.Max(latMax, v[k][1])

				lngMin = math.Min(lngMin, v[k][0])
				lngMax = math.Max(lngMax, v[k][0])
			}
		}

	case GEOJSON_MULTI_POINT:
		for k, v := range e.Features[e.setOfFeatures].Geometry.Coordinates.([]multiPoint) {
			if k == 0 {
				latMin = v[1]
				latMax = v[1]

				lngMin = v[0]
				lngMax = v[0]
			} else {
				latMin = math.Min(latMin, v[1])
				latMax = math.Max(latMax, v[1])

				lngMin = math.Min(lngMin, v[0])
				lngMax = math.Max(lngMax, v[0])
			}
		}

	case GEOJSON_MULTI_LINE_STRING:
		for k, v := range e.Features[e.setOfFeatures].Geometry.Coordinates.([]multiLineString) {
			if k == 0 {
				latMin = v[k][1]
				latMax = v[k][1]

				lngMin = v[k][0]
				lngMax = v[k][0]
			} else {
				latMin = math.Min(latMin, v[k][1])
				latMax = math.Max(latMax, v[k][1])

				lngMin = math.Min(lngMin, v[k][0])
				lngMax = math.Max(lngMax, v[k][0])
			}
		}

	case GEOJSON_MULTI_POLYGON:
		for kmp, vmp := range e.Features[e.setOfFeatures].Geometry.Coordinates.([]multiPolygon) {
			for kp, vp := range vmp {
				for k, v := range vp {
					if k == 0 && kmp == 0 && kp == 0 {
						latMin = v[1]
						latMax = v[1]

						lngMin = v[0]
						lngMax = v[0]
					} else {
						latMin = math.Min(latMin, v[1])
						latMax = math.Max(latMax, v[1])

						lngMin = math.Min(lngMin, v[0])
						lngMax = math.Max(lngMax, v[0])
					}
				}
			}
		}

	}

	e.Features[e.setOfFeatures].Geometry.BoundingBox = [4]float64{lngMin, latMin, lngMax, latMax}
}

func (e *GeoJSon) AddLatLng(lat, lng float64) {
	e.AddLatLngAlt(lat, lng, 0.0)
}

func (e *GeoJSon) AddLngLat(lng, lat float64) {
	e.AddLatLngAlt(lat, lng, 0.0)
}

func (e *GeoJSon) AddLngLatAlt(lng, lat, alt float64) {
	e.AddLatLngAlt(lat, lng, alt)
}

func (e *GeoJSon) AddLatLngAlt(lat, lng, alt float64) {
	switch e.Features[e.setOfFeatures].Geometry.typeConst {
	case GEOJSON_POINT:
		e.Features[e.setOfFeatures].Geometry.Coordinates = point{}
		e.Features[e.setOfFeatures].Geometry.Coordinates = [3]float64{lng, lat, alt}

	case GEOJSON_LINE_STRING:
		if len(e.Features[e.setOfFeatures].Geometry.Coordinates.([]lineString)) == 0 {
			e.Features[e.setOfFeatures].Geometry.Coordinates = make([]lineString, 0)
		}

		e.Features[e.setOfFeatures].Geometry.Coordinates = append(e.Features[e.setOfFeatures].Geometry.Coordinates.([]lineString), [3]float64{lng, lat, alt})

	case GEOJSON_POLYGON:
		if len(e.Features[e.setOfFeatures].Geometry.Coordinates.([]polygon)) == 0 {
			e.Features[e.setOfFeatures].Geometry.Coordinates = make([]polygon, 1)
		}

		if len(e.Features[e.setOfFeatures].Geometry.Coordinates.([]polygon)[e.Features[e.setOfFeatures].setOfCoordinates]) == 0 {
			e.Features[e.setOfFeatures].Geometry.Coordinates.([]polygon)[e.Features[e.setOfFeatures].setOfCoordinates] = make(polygon, 0)
		}

		e.Features[e.setOfFeatures].Geometry.Coordinates.([]polygon)[e.Features[e.setOfFeatures].setOfCoordinates] = append(e.Features[e.setOfFeatures].Geometry.Coordinates.([]polygon)[e.Features[e.setOfFeatures].setOfCoordinates], [3]float64{lng, lat, alt})

	case GEOJSON_MULTI_POINT:
		if len(e.Features[e.setOfFeatures].Geometry.Coordinates.([]multiPoint)) == 0 {
			e.Features[e.setOfFeatures].Geometry.Coordinates = make([]multiPoint, 0)
		}

		e.Features[e.setOfFeatures].Geometry.Coordinates = append(e.Features[e.setOfFeatures].Geometry.Coordinates.([]multiPoint), [3]float64{lng, lat, alt})

	case GEOJSON_MULTI_LINE_STRING:
		if len(e.Features[e.setOfFeatures].Geometry.Coordinates.([]multiLineString)) == 0 {
			e.Features[e.setOfFeatures].Geometry.Coordinates = make([]multiLineString, 1)
		}

		if len(e.Features[e.setOfFeatures].Geometry.Coordinates.([]multiLineString)[e.Features[e.setOfFeatures].setOfCoordinates]) == 0 {
			e.Features[e.setOfFeatures].Geometry.Coordinates.([]multiLineString)[e.Features[e.setOfFeatures].setOfCoordinates] = make(multiLineString, 0)
		}

		e.Features[e.setOfFeatures].Geometry.Coordinates.([]multiLineString)[e.Features[e.setOfFeatures].setOfCoordinates] = append(e.Features[e.setOfFeatures].Geometry.Coordinates.([]multiLineString)[e.Features[e.setOfFeatures].setOfCoordinates], [3]float64{lng, lat, alt})

	case GEOJSON_MULTI_POLYGON:
		if len(e.Features[e.setOfFeatures].Geometry.Coordinates.([]multiPolygon)) == 0 {
			e.Features[e.setOfFeatures].Geometry.Coordinates = make([]multiPolygon, 1)
		}

		if len(e.Features[e.setOfFeatures].Geometry.Coordinates.([]multiPolygon)[e.Features[e.setOfFeatures].setOfCoordinates]) == 0 {
			e.Features[e.setOfFeatures].Geometry.Coordinates.([]multiPolygon)[e.Features[e.setOfFeatures].setOfCoordinates] = make(multiPolygon, e.Features[e.setOfFeatures].setOfPolygons)
		}

		e.Features[e.setOfFeatures].Geometry.Coordinates.([]multiPolygon)[e.Features[e.setOfFeatures].setOfCoordinates][e.Features[e.setOfFeatures].setOfLines] = append(e.Features[e.setOfFeatures].Geometry.Coordinates.([]multiPolygon)[e.Features[e.setOfFeatures].setOfCoordinates][e.Features[e.setOfFeatures].setOfLines], [3]float64{lng, lat, alt})

	}
}

func (e *GeoJSon) String() (string, error) {
	byteJSon, err := json.Marshal(e)

	return string(byteJSon), err
}

func (e *GeoJSon) StringLastFeature() (string, error) {
	byteJSon, err := json.Marshal(e.Features[len(e.Features)-1])

	return string(byteJSon), err
}

func (e *GeoJSon) StringAllFeatures() (string, error) {
	var features string = ""
	var byteJSon []byte
	var err error

	byteJSon, err = json.Marshal(e.Features[len(e.Features)-1])

	for k, feature := range e.Features {
		if k != 0 {
			features += ","
		}

		byteJSon, err = json.Marshal(feature)
		if err != nil {
			return "", err
		}

		features += string(byteJSon)
	}

	return features, err
}
