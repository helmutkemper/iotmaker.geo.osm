package iotmaker_geo_osm

import (
	"fmt"
	"github.com/helmutkemper/gOsm/consts"
)

type AngleStt struct {
	Degrees      float64 //Angle
	Radians      float64 //Angle
	Unit         string  //Unit
	PreserveUnit string  //original Unit
}

type AngleListStt struct {
	List []AngleStt
}

// Set Angle value as decimal degrees
func (a *AngleStt) SetDecimalDegrees(degreesAFlt, primesAFlt, secondsAFlt float64) {
	a.Degrees = degreesAFlt + primesAFlt/60.0 + secondsAFlt/3600.0
	a.Radians = DegreesToRadians(degreesAFlt + primesAFlt/60.0 + secondsAFlt/3600.0)
	a.Unit = consts.DEGREES
	a.PreserveUnit = consts.DEGREES
}

// Set Angle value as degrees
func (a *AngleStt) SetDegrees(angleAFlt float64) {
	a.Degrees = angleAFlt
	a.Radians = DegreesToRadians(angleAFlt)
	a.Unit = consts.DEGREES
	a.PreserveUnit = consts.DEGREES
}

// Set Angle value as radians
func (a *AngleStt) SetRadians(angleAFlt float64) {
	a.Radians = angleAFlt
	a.Degrees = RadiansToDegrees(angleAFlt)
	a.Unit = consts.RADIANS
	a.PreserveUnit = consts.RADIANS
}

// Set Angle value as degrees
func (a *AngleStt) AddDegrees(angleAFlt float64) {
	a.Degrees = a.Degrees + angleAFlt
	a.Radians = DegreesToRadians(a.Degrees)
}

// Get Angle
func (a *AngleStt) GetAsRadians() float64 {
	return a.Radians
}

// Get Angle
func (a *AngleStt) GetAsDegrees() float64 {
	return a.Degrees
}

// Get Unit
func (a *AngleStt) GetUnit() string {
	return a.Unit
}

// Get original Unit before conversion
func (a *AngleStt) GetOriginalUnit() string {
	return a.PreserveUnit
}

// Get Angle as string
func (a *AngleStt) ToDegreesString() string {
	return fmt.Sprintf("%1.3f%v", a.Degrees, consts.DEGREES)
}

// Get Angle as string
func (a *AngleStt) ToRadiansString() string {
	return fmt.Sprintf("%1.3f%v", a.Radians, consts.RADIANS)
}
