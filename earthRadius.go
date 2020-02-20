package iotmaker_geo_osm

import (
	"github.com/helmutkemper/gOsm/consts"
	"math"
)

// Earth radius at a given latitude, according to the GEOIDAL CONST ellipsoid, in meters
func EarthRadius(pointAStt PointStt) DistanceStt {
	latitudeLFlt := pointAStt.Rad[1]

	var returnLStt DistanceStt

	returnLStt.SetMeters(
		math.Sqrt(
			(math.Pow(math.Pow(consts.GEOIDAL_MAJOR, 2.0)*math.Cos(latitudeLFlt), 2.0) +
				math.Pow(math.Pow(consts.GEOIDAL_MINOR, 2.0)*math.Sin(latitudeLFlt), 2.0)) /

				(math.Pow(consts.GEOIDAL_MAJOR*math.Cos(latitudeLFlt), 2.0) +
					math.Pow(consts.GEOIDAL_MINOR*math.Sin(latitudeLFlt), 2.0))))

	return returnLStt
}
