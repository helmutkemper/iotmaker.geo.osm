package iotmaker_geo_osm

import (
	"math"
)

func DegreesToRadians(a float64) float64 { return math.Pi * a / 180.0 }

func RadiansToDegrees(a float64) float64 { return 180.0 * a / math.Pi }

func Pythagoras(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt(math.Pow(x2-x1, 2.0) + math.Pow(y2-y1, 2.0))
}
