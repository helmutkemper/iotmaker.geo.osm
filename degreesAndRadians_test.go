package iotmaker_geo_osm

import "fmt"

func ExampleDegreesToRadians() {
	fmt.Printf("Radians: %v\n", DegreesToRadians(180.0))

	// Output:
	// Radians: 3.141592653589793
}

func ExampleDegreesToRadians2() {
	fmt.Printf("Degrees: %v\n", RadiansToDegrees(3.14159265358979323))

	// Output:
	// Degrees: 180
}
