package iotmaker_geo_osm

import (
	"fmt"
	"github.com/helmutkemper/gOsm/consts"
)

type BoxStt struct {
	BottomLeft PointStt
	UpperRight PointStt
}

type BoxListStt struct {
	List []BoxStt
}

// Get unit
// Get angle as string
func (boxAStt *BoxStt) ToDegreesString() string {
	return fmt.Sprintf("((%1.5f,%1.5f),(%1.5f,%1.5f))%v", boxAStt.BottomLeft.Loc[0], boxAStt.BottomLeft.Loc[1], boxAStt.UpperRight.Loc[0], boxAStt.UpperRight.Loc[1], consts.DEGREES)
}

// Get angle as string
func (boxAStt *BoxStt) ToRadiansString() string {
	return fmt.Sprintf("((%1.5f,%1.5f),(%1.5f,%1.5f))%v", boxAStt.BottomLeft.Rad[0], boxAStt.BottomLeft.Rad[1], boxAStt.UpperRight.Rad[0], boxAStt.UpperRight.Rad[1], consts.RADIANS)
}

func (boxAStt *BoxStt) ToGoogleMapString() string {
	return fmt.Sprintf("BottomLeft: %1.5f, %1.5f [ Please, copy and past this value on google maps search ]\nUpperRight: %1.5f, %1.5f [ Please, copy and past this value on google maps search ]", boxAStt.BottomLeft.Loc[1], boxAStt.BottomLeft.Loc[0], boxAStt.UpperRight.Loc[1], boxAStt.UpperRight.Loc[0])
}

// Bounding box surrounding the point at given coordinates,
func (boxAStt *BoxStt) Make(pointAStt PointStt, distanceAStt DistanceStt) {
	boxLStt := BoundingBox(pointAStt, distanceAStt)

	boxAStt.BottomLeft = boxLStt.BottomLeft
	boxAStt.UpperRight = boxLStt.UpperRight
}
