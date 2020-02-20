package iotmaker_geo_osm

import (
	"github.com/helmutkemper/mgo/bson"
	"math"
)

// en: Returns a box that is compatible with the perimeter of the object.
//
// For the better performance of the database, never look for the points contained within a radius, look for the points contained within a rectangular box with the function $box of MongoDB
//
// Returns the answer in degrees
//
// pt: Devolve uma caixa compatível com o perímetro do objeto.
//
// Para melhor desempenho do banco de dados, nunca procure pontos contidos dentro de um raio, procure pontos contidos dentro de uma caixa retangular com a função $box do MongoDB
//
// Devolve a resposta em graus decimais
func GetBox(list *[]PointStt) BoxStt {
	var latMin, latMax, lngMin, lngMax float64

	for k, v := range *list {
		if k == 0 {
			latMin = v.GetLatitudeAsRadians()
			latMax = v.GetLatitudeAsRadians()

			lngMin = v.GetLongitudeAsRadians()
			lngMax = v.GetLongitudeAsRadians()
		} else {
			latMin = math.Min(latMin, v.GetLatitudeAsRadians())
			latMax = math.Max(latMax, v.GetLatitudeAsRadians())

			lngMin = math.Min(lngMin, v.GetLongitudeAsRadians())
			lngMax = math.Max(lngMax, v.GetLongitudeAsRadians())
		}
	}

	boxLStt := BoxStt{}
	boxLStt.UpperRight.SetLngLatRadians(lngMin, latMax)
	boxLStt.BottomLeft.SetLngLatRadians(lngMax, latMin)

	return boxLStt
}

// en: Returns a box that is compatible with the function $box of MongoDb.
//
// For the better performance of the database, never look for the points contained within a radius, search for the points contained within a rectangular box with the function $box of MongoDB.
//
// Returns the answer in decimal degrees
//
// pt: Devolve uma caixa compatível com o perímetro do objeto no formato bson e é compatível com a função $box do MongoDB.
//
// Para melhor desempenho do banco de dados, nunca procure pontos contidos dentro de um raio, procure pontos contidos dentro de uma caixa retangular com a função $box do MongoDB.
//
// Devolve a resposta em graus decimais.
func GetBSonBoxInDegrees(list *[]PointStt) bson.M {
	boxLStt := GetBox(list)
	return bson.M{"$geoWithin": bson.M{"$box": [2][2]float64{boxLStt.BottomLeft.Loc, boxLStt.UpperRight.Loc}}}
}

// en: Returns a box that is compatible with the perimeter of the object.
//
// For the better performance of the database, never look for the points contained within a radius, look for the points contained within a rectangular box with the function $box of MongoDB.
//
// Returns the answer in degrees.
//
// pt: Devolve uma caixa compatível com o perímetro do objeto.
//
// Para melhor desempenho do banco de dados, nunca procure pontos contidos dentro de um raio, procure pontos contidos dentro de uma caixa retangular com a função $box do MongoDB.
//
// Devolve a resposta em graus decimai
func GetBoxFlt(list *[][2]float64) BoxStt {
	var latMin, latMax, lngMin, lngMax float64

	for k, v := range *list {
		if k == 0 {
			latMin = v[1]
			latMax = v[1]

			lngMin = v[0]
			lngMax = v[0]
		} else {
			latMin = math.Min(latMin, v[1])
			latMax = math.Max(latMax, v[1])

			lngMin = math.Min(lngMin, v[0])
			lngMax = math.Max(lngMax, v[0])
		}
	}

	boxLStt := BoxStt{}
	boxLStt.BottomLeft.SetLngLatRadians(lngMin, latMax)
	boxLStt.UpperRight.SetLngLatRadians(lngMax, latMin)

	return boxLStt
}

// en: Returns a box that is compatible with the function $box of MongoDb.
//
// For the better performance of the database, never look for the points contained within a radius, search for the points contained within a rectangular box with the function $box of MongoDB.
//
// Returns the answer in decimal degrees.
//
// pt: Devolve uma caixa compatível com o perímetro do objeto no formato bson e é compatível com a função $box do MongoDB.
//
// Para melhor desempenho do banco de dados, nunca procure pontos contidos dentro de um raio, procure pontos contidos dentro de uma caixa retangular com a função $box do MongoDB.
//
// Devolve a resposta em graus decimais.
func GetBSonBoxInDegreesFlt(list *[][2]float64) bson.M {
	boxLStt := GetBoxFlt(list)
	return bson.M{"$geoWithin": bson.M{"$box": [2][2]float64{boxLStt.BottomLeft.Loc, boxLStt.UpperRight.Loc}}}
}

// en: Returns a box that is compatible with the perimeter of the object.
//
// For the better performance of the database, never look for the points contained within a radius, look for the points contained within a rectangular box with the function $box of MongoDB.
//
// Returns the answer in degrees.
//
// pt: Devolve uma caixa compatível com o perímetro do objeto.
//
// Para melhor desempenho do banco de dados, nunca procure pontos contidos dentro de um raio, procure pontos contidos dentro de uma caixa retangular com a função $box do MongoDB.
//
// Devolve a resposta em graus decimais.
func GetBoxList(list *[]PointListStt) BoxStt {
	var latMin, latMax, lngMin, lngMax float64

	for sk, subList := range *list {
		for k, v := range subList.List {
			if k == 0 && sk == 0 {
				latMin = v.GetLatitudeAsRadians()
				latMax = v.GetLatitudeAsRadians()

				lngMin = v.GetLatitudeAsRadians()
				lngMax = v.GetLatitudeAsRadians()
			} else {
				latMin = math.Min(latMin, v.GetLatitudeAsRadians())
				latMax = math.Max(latMax, v.GetLatitudeAsRadians())

				lngMin = math.Min(lngMin, v.GetLatitudeAsRadians())
				lngMax = math.Max(lngMax, v.GetLatitudeAsRadians())
			}
		}
	}

	boxLStt := BoxStt{}
	boxLStt.BottomLeft.SetLngLatRadians(lngMin, latMax)
	boxLStt.UpperRight.SetLngLatRadians(lngMax, latMin)

	return boxLStt
}

// en: Returns a box that is compatible with the function $box of MongoDb.
//
// For the better performance of the database, never look for the points contained within a radius, search for the points contained within a rectangular box with the function $box of MongoDB.
//
// Returns the answer in decimal degrees.
//
// pt: Devolve uma caixa compatível com o perímetro do objeto no formato bson e é compatível com a função $box do MongoDB.
//
// Para melhor desempenho do banco de dados, nunca procure pontos contidos dentro de um raio, procure pontos contidos dentro de uma caixa retangular com a função $box do MongoDB.
//
// Devolve a resposta em graus decimais.
func GetBSonBoxInDegreesList(list *[]PointListStt) bson.M {
	boxLStt := GetBoxList(list)
	return bson.M{"$geoWithin": bson.M{"$box": [2][2]float64{boxLStt.BottomLeft.Loc, boxLStt.UpperRight.Loc}}}
}

// en: Returns a box that is compatible with the perimeter of the object.
//
// For the better performance of the database, never look for the points contained within a radius, look for the points contained within a rectangular box with the function $box of MongoDB.
//
// Returns the answer in degrees.
//
// pt: Devolve uma caixa compatível com o perímetro do objeto.
//
// Para melhor desempenho do banco de dados, nunca procure pontos contidos dentro de um raio, procure pontos contidos dentro de uma caixa retangular com a função $box do MongoDB.
//
// Devolve a resposta em graus decimais
func GetBoxPolygonList(list *PolygonListStt) BoxStt {
	var latMin, latMax, lngMin, lngMax float64

	for sk, subList := range list.List {
		for k, v := range subList.PointsList {
			if k == 0 && sk == 0 {
				latMin = v.GetLatitudeAsRadians()
				latMax = v.GetLatitudeAsRadians()

				lngMin = v.GetLatitudeAsRadians()
				lngMax = v.GetLatitudeAsRadians()
			} else {
				latMin = math.Min(latMin, v.GetLatitudeAsRadians())
				latMax = math.Max(latMax, v.GetLatitudeAsRadians())

				lngMin = math.Min(lngMin, v.GetLatitudeAsRadians())
				lngMax = math.Max(lngMax, v.GetLatitudeAsRadians())
			}
		}
	}

	boxLStt := BoxStt{}
	boxLStt.BottomLeft.SetLngLatRadians(lngMin, latMax)
	boxLStt.UpperRight.SetLngLatRadians(lngMax, latMin)

	return boxLStt
}

// en: Returns a box that is compatible with the function $box of MongoDb.
//
// For the better performance of the database, never look for the points contained within a radius, search for the points contained within a rectangular box with the function $box of MongoDB.
//
// Returns the answer in decimal degrees.
//
// pt: Devolve uma caixa compatível com o perímetro do objeto no formato bson e é compatível com a função $box do MongoDB.
//
// Para melhor desempenho do banco de dados, nunca procure pontos contidos dentro de um raio, procure pontos contidos dentro de uma caixa retangular com a função $box do MongoDB.
//
// Devolve a resposta em graus decimais.
func GetBSonBoxInDegreesPolygonList(list *PolygonListStt) bson.M {
	boxLStt := GetBoxPolygonList(list)
	return bson.M{"$geoWithin": bson.M{"$box": [2][2]float64{boxLStt.BottomLeft.Loc, boxLStt.UpperRight.Loc}}}
}
