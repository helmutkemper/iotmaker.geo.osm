package iotmaker_geo_osm

import (
	"fmt"
	"math"
)

//todo unidades para constantes
type DistanceStt struct {
	Meters       float64 // distance
	Kilometers   float64 // distance
	unit         string  // distance unit
	preserveUnit string  // original unit
}

type DistanceListStt struct {
	List []DistanceStt
}

// Get distance value
func (d *DistanceStt) GetMeters() float64 {
	return d.Meters
}

func (d *DistanceStt) GetKilometers() float64 {
	return d.Kilometers
}

// Get distance unit
func (d *DistanceStt) GetUnit() string {
	return d.unit
}

func (d *DistanceStt) GetOriginalUnit() string {
	return d.preserveUnit
}

// Set distance as meters
func (d *DistanceStt) AddMeters(m float64) {

	d.Meters += m
	d.Kilometers += m / 1000
}

// Set distance as meters
func (d *DistanceStt) SetMeters(m float64) {
	d.Meters = m
	d.Kilometers = m / 1000
	d.unit = "m"
	d.preserveUnit = "m"
}

func (d *DistanceStt) SetMetersIfGreaterThan(m float64) {
	test := math.Max(d.Meters, m)

	d.Meters = test
	d.Kilometers = test / 1000
	d.unit = "m"
	d.preserveUnit = "m"
}

func (d *DistanceStt) SetKilometersIfGreaterThan(km float64) {
	test := math.Max(d.Kilometers, km)

	d.Meters = test * 1000
	d.Kilometers = test
	d.unit = "km"
	d.preserveUnit = "km"
}

func (d *DistanceStt) SetMetersIfLessThan(m float64) {
	test := math.Min(d.Meters, m)

	d.Meters = test
	d.Kilometers = test / 1000
	d.unit = "m"
	d.preserveUnit = "m"
}

func (d *DistanceStt) SetKilometersIfLessThan(km float64) {
	test := math.Min(d.Kilometers, km)

	d.Meters = test * 1000
	d.Kilometers = test
	d.unit = "km"
	d.preserveUnit = "km"
}

// Set distance as kilometers
func (d *DistanceStt) AddKilometers(km float64) {
	d.Meters += km * 1000
	d.Kilometers += km
	d.unit = "Km"
	d.preserveUnit = "Km"
}

// Set distance as kilometers
func (d *DistanceStt) SetKilometers(km float64) {
	d.Meters = km * 1000
	d.Kilometers = km
	d.unit = "Km"
	d.preserveUnit = "Km"
}

// Get distance as string
func (d *DistanceStt) ToMetersString() string {
	return fmt.Sprintf("%1.2fm", d.Meters)
}

func (d *DistanceStt) ToKilometersString() string {
	return fmt.Sprintf("%1.2fKm", d.Kilometers)
}
