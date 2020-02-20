package iotmaker_geo_osm

import (
	"math"
)

// Calculate distance between two points
func DistanceBetweenTwoPoints(pointAAStt, pointBAStt PointStt) DistanceStt {
	var returnLStt DistanceStt

	earthRadiusA := EarthRadius(pointAAStt)

	//dist = arccos(sin(lat1) · sin(lat2) + cos(lat1) · cos(lat2) · cos(lon1 - lon2)) · R
	distLFlt := math.Acos(math.Sin(pointAAStt.Rad[1])*math.Sin(pointBAStt.Rad[1])+
		math.Cos(pointAAStt.Rad[1])*math.Cos(pointBAStt.Rad[1])*
			math.Cos(pointAAStt.Rad[0]-pointBAStt.Rad[0])) *
		earthRadiusA.GetKilometers()

	if math.IsNaN(distLFlt) {
		distLFlt = 0
	}

	returnLStt.SetKilometers(distLFlt)

	return returnLStt
}
