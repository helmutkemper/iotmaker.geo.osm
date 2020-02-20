package iotmaker_geo_osm

import (
	"math"
)

// x: longitude
// y: latitude

// Calculate new point at given distance and angle
func DestinationPoint(pointAStt PointStt, distanceAStt DistanceStt, angleAStt AngleStt) PointStt {
	var returnLStt PointStt

	earthRadius := EarthRadius(pointAStt)

	//latB = asin( sin( latA ) * cos( d / R ) +
	//  cos( latA ) * sin( d / R ) * cos( θ ) )
	y := math.Asin(math.Sin(pointAStt.Rad[1])*math.Cos(distanceAStt.GetMeters()/earthRadius.GetMeters()) +
		math.Cos(pointAStt.Rad[1])*math.Sin(distanceAStt.GetMeters()/earthRadius.GetMeters())*math.Cos(angleAStt.GetAsRadians()))

	//lonB = lonA + atan2( sin( θ ) *
	//  sin( d / R ) * cos( latA ),
	//  cos( d / R ) − sin( latA ) * sin( latB ) )
	x := pointAStt.Rad[0] + math.Atan2(math.Sin(angleAStt.GetAsRadians())*
		math.Sin(distanceAStt.GetMeters()/earthRadius.GetMeters())*math.Cos(pointAStt.Rad[1]),
		math.Cos(distanceAStt.GetMeters()/earthRadius.GetMeters())-math.Sin(pointAStt.Rad[1])*math.Sin(y))

	returnLStt.SetXYRadians(x, y)

	return returnLStt
}
