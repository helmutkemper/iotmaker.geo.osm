package iotmaker_geo_osm

import (
	"math"
)

// Calculate angle between two points
func DirectionBetweenTwoPoints(pointAAStt, pointBAStt PointStt) AngleStt {
	var directionLFlt AngleStt

	//dlat := pointBAStt.GetLatitudeAsRadians - pointAAStt.GetLatitudeAsRadians
	//dlon := pointBAStt.GetLongitudeAsRadians - pointAAStt.GetLongitudeAsRadians

	y := math.Sin(pointBAStt.GetLongitudeAsRadians()-pointAAStt.GetLongitudeAsRadians()) * math.Cos(pointBAStt.GetLatitudeAsRadians())
	x := math.Cos(pointAAStt.GetLatitudeAsRadians())*math.Sin(pointBAStt.GetLatitudeAsRadians()) - math.Sin(pointAAStt.GetLatitudeAsRadians())*math.Cos(pointBAStt.GetLatitudeAsRadians())*math.Cos(pointBAStt.GetLongitudeAsRadians()-pointAAStt.GetLongitudeAsRadians())
	if y > 0.0 {
		if x > 0.0 {
			directionLFlt.SetRadians(math.Atan(y / x))
		}
		if x < 0.0 {
			directionLFlt.SetRadians(math.Pi - math.Atan(-y/x))
		}
		if x == 0.0 {
			directionLFlt.SetRadians(math.Pi / 2.0)
		}
	}
	if y < 0.0 {
		if x > 0.0 {
			directionLFlt.SetRadians(-math.Atan(-y/x) + 2.0*math.Pi)
		}
		if x < 0.0 {
			directionLFlt.SetRadians(math.Atan(y/x) - math.Pi + 2.0*math.Pi) //ok
		}
		if x == 0.0 {
			directionLFlt.SetRadians(math.Pi * 3.0 / 2.0)
		}
	}
	if y == 0.0 {
		if x > 0.0 {
			directionLFlt.SetRadians(0.0)
		}
		if x < 0.0 {
			directionLFlt.SetRadians(math.Pi)
		}
		if x == 0.0 {
			directionLFlt.SetRadians(0.0)
		}
	}

	/*
	    angleLFlt := math.Atan2( pointBAStt.GetLatitudeAsRadians - pointAAStt.GetLatitudeAsRadians, pointBAStt.GetLongitudeAsRadians - pointAAStt.GetLongitudeAsRadians )
	  	angleLFlt -= ( math.Pi / 2 ) * 1.226949
	    directionLFlt.SetRadians( angleLFlt )
	*/

	return directionLFlt
}
