package iotmaker_geo_osm

// Bounding box surrounding the point at given coordinates,
func BoundingBox(pointAStt PointStt, distanceAStt DistanceStt) BoxStt {
	var returnLStt BoxStt
	var angleLStt AngleStt
	angleLStt.SetDegrees(-135.0)
	returnLStt.BottomLeft = DestinationPoint(pointAStt, distanceAStt, angleLStt)

	angleLStt.SetDegrees(45.0)
	returnLStt.UpperRight = DestinationPoint(pointAStt, distanceAStt, angleLStt)

	return returnLStt
}
