package iotmaker_geo_osm

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"github.com/helmutkemper/gOsm/utilMath"
	"github.com/helmutkemper/mgo/bson"
	log "github.com/helmutkemper/seelog"
	"github.com/helmutkemper/zstd"
	"io"
	"io/ioutil"
	"math"
	"os"
	"strconv"
)

/*
{
  "bbox.bottomleft.loc.0": { $gte: -55.799560546875 },
  "bbox.bottomleft.loc.1": { $lte: -0.00549316405408165 },
  "bbox.upperright.loc.0": { $lte: -55.799560546875 },
  "bbox.upperright.loc.1": { $gte: -0.00549316405408165 }
}

*/

// English: Polygons are mainly used to demarcate boundaries, and to allow searches in limited areas.
//
// In OpenStreetMaps there is not much difference between a polygon and a way, being a polygon, a way where the first and last point are repeated.
//
// In our case, a polygon can be assembled from a single way, or from the concatenation of several ways distinct.
//
// Português: Polígonos são usados principalmente para demarcar fronteiras e permitir buscas em áreas limitadas.
//
// No OpenStreetMaps não há muita diferença entre um polígono e um way, sendo um polígono um way onde o primeiro e último ponto são repetidos.
//
// No nosso caso, um polígono pode ser montado a partir de um único way ou a partir da concatenação de vários ways distintos.
type PolygonStt struct {

	// English: id open street maps
	//
	// Português: id do open street maps
	Id int64 `bson:"id"`

	Surrounding float64 `bson:"surrounding"`

	Visible bool `bson:"bool"`

	// English: Tags OpenStreetMaps
	//
	// The Tags contain all kinds of information, as long as they were imported, the name of a commercial establishment, for example.
	//
	// English: Tags do Open Street Maps
	//
	// As Tags contêm todo tipo de informação, desde como elas foram importadas, ao nome de um estabelecimento comercial, por exemplo.
	Tag map[string]string `bson:"tag"`

	TagFromWay map[string]map[string]string `bson:"tagFromWay"`

	// English: due to some driver mgo limitations, gives less problem if the tags very large, they are separate. By this, the entire content of
	//
	// the tag starts with 'name:.*' it is played within the 'international' key
	//
	// Português: devido a algumas limitações do driver mgo, dá menos problema se as tags muito grandes forem separadas. Por isto, todo conteúdo de
	//
	// tag iniciado com 'name:.*' é jogado dentro da chave 'International'
	International map[string]string `bson:"international"`

	// English: user data
	//
	// Português: dados do usuário
	Data map[string]string `bson:"data"`

	// English: List of polygon forming points
	//
	// Português: Lista dos pontos formadores do polígono
	PointsList []PointStt `bson:"pointList"`

	// English: The amount of forming points of the polygon
	//
	// Português: Quantidade de pontos formadores do polígono
	Length int `bson:"length"`

	// English: The area of the polygon used for the calculation of the centroid. I do not recommend the use for calculation of the geographic area in this version.
	//
	// Português: Área do polígono usada para o calculo da centroide. Não recomendo o uso para calculo de área geográfica nessa versão.
	Area float64 `bson:"area"`

	// English: Centroid of polygon
	//
	// Português: Centroide do polígono
	Centroid PointStt `bson:"centroid"`

	// English: used to test if the polygon has been initialized
	//
	// Português: usado para testar se polígono foi inicializado
	Initialize bool `bson:"inicialize"`

	// English: used in the calculations of the point inside the polygon.
	//
	// should not be changed manually
	//
	// Português: usado nos cálculos do ponto dentro do polígono.
	//
	// não deve ser alterado manualmente
	ConstantMAFlt []float64 `bson:"constantMAFlt"`

	// English: used in the calculations of the point inside the polygon.
	//
	// should not be changed manually
	//
	// Português: usado nos cálculos do ponto dentro do polígono.
	//
	// não deve ser alterado manualmente
	MultipleMAFlt []float64 `bson:"multipleMAFlt"`

	// English: distance to the nearest point of the perimeter in the order that points were added
	//
	// Português: distância para o próximo ponto do perímetro na ordem que os pontos foram adicionados
	Distance []DistanceStt `bson:"distance"`

	// English: the total length of the perimeter
	//
	// Português: comprimento total do perímetro
	DistanceTotal DistanceStt `bson:"distanceTotal"`

	// English: ângulo em relação ao próximo ponto do perímetro na ordem que os pontos foram adicionados
	//
	// Português: angle in relation to the next point of the perimeter in the order that points were added
	Angle []AngleStt `bson:"angle"`

	// English: boundary box in degrees
	//
	// Português: caixa de perímetro em graus decimais
	BBox BoxStt `bson:"bbox"`

	/*
	   bson.M{ "bBoxSearch.1.0": bson.M{ "$gte": -62.162704467 }, "bBoxSearch.1.1": bson.M{ "$lte": -12.341343394 }, "bBoxSearch.0.0": bson.M{ "$lte": -62.162704467 }, "bBoxSearch.0.1": bson.M{ "$gte": -12.341343394 } }
	*/
	//BBoxSearch [2][2]float64 `bson:"bBoxSearch"`

	// English: boundary box in BSon to MongoDB
	//
	// Português: caixa de perímetro em BSon para o MongoDB
	//BBoxBSon       bson.M   `bson:"bBoxBSon"`

	GeoJSonFeature string   `bson:"geoJSonFeature"`
	tmp            []WayStt `bson:"-" json:"-"`

	IdWay       []int64         `bson:"idWay"`
	idWayUnique map[int64]int64 `bson:"-"`

	MinimalDistance DistanceStt `bson:"MinimalDistance" json:"-"`

	Md5  [16]byte `bson:"md5" json:"-"`
	Size int      `bson:"size" json:"-"`

	HasKeyValue bool `bson:"hasKeyValue" json:"-"`
}

func (el *PolygonStt) GetIdAsByte() []byte {
	var ret = make([]byte, 8)
	binary.LittleEndian.PutUint64(ret, uint64(el.Id))

	return ret
}

func (el *PolygonStt) MakeMD5() (error, []byte) {
	var err error
	var byteBSon []byte

	el.Size = 0
	el.Md5 = [16]byte{}
	byteBSon, err = bson.Marshal(el)
	if err != nil {
		log.Criticalf("gOsm.geoMath.tmpPoint.error: %s", err.Error())
		return err, []byte{}
	}

	el.Md5 = md5.Sum(byteBSon)
	el.Size = len(byteBSon)

	byteBSon, err = bson.Marshal(el)
	if err != nil {
		log.Criticalf("gOsm.geoMath.tmpPoint.error: %s", err.Error())
		return err, []byte{}
	}

	return nil, byteBSon
}

func (el *PolygonStt) ToBSon() (error, []byte) {
	return el.MakeMD5()
}

func (el *PolygonStt) FromBSon(byteBSon []byte) error {
	var err error

	err = bson.Unmarshal(byteBSon, el)
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err)
	}

	return err
}

func (el *PolygonStt) AsArray() []PolygonStt {
	var returnLStt []PolygonStt = make([]PolygonStt, 1)
	returnLStt[0] = *el

	return returnLStt
}

func (el *PolygonStt) SetMinimalDistance(distanceAStt DistanceStt) {
	el.MinimalDistance = distanceAStt
}

func (el *PolygonStt) SetSurrounding(distance float64) {
	el.Surrounding = distance
}

// English: Copies the data of the relation in the polygon.
//
// Português: Copia os dados de uma relação no polígono.
//
// @see dataOfOsm in blog.osm.io
func (el *PolygonStt) AddRelationDataAsPolygonData(relation *RelationStt) {
	el.Visible = (*relation).Visible

	el.Tag = (*relation).Tag
	el.International = (*relation).International
}

// English: Copies the data of the way in the polygon.
//
// Português: Copia os dados de uma way no polígono.
//
// @see dataOfOsm in blog.osm.io
func (el *PolygonStt) AddWayDataAsPolygonData(way *WayStt) {
	if len(el.TagFromWay) == 0 {
		el.TagFromWay = make(map[string]map[string]string)
	}

	el.TagFromWay[strconv.FormatInt((*way).Id, 10)] = (*way).Tag

	if len(el.IdWay) == 0 {
		el.IdWay = make([]int64, 0)
	}

	if len(el.idWayUnique) == 0 {
		el.idWayUnique = make(map[int64]int64)
	}
	if el.idWayUnique[(*way).Id] != (*way).Id {
		el.idWayUnique[(*way).Id] = (*way).Id
		el.IdWay = append(el.IdWay, (*way).Id)
	}

	//el.Id             =  (*way).Id

	//mapTagLock.Lock()
	//el.Tag            =  (*way).Tag
	//mapTagLock.Unlock()

	el.International = (*way).International

	el.Data = (*way).Data
}

// English: Turn a way in a polygon.
//
// Português: Transforma um way em um polígono.
func (el *PolygonStt) AddWayAsPolygon(way *WayStt) {
	if len(el.TagFromWay) == 0 {
		el.TagFromWay = make(map[string]map[string]string)
	}

	if len((*way).Tag) > 0 {
		el.TagFromWay[strconv.FormatInt((*way).Id, 10)] = (*way).Tag
	}

	if len(el.IdWay) == 0 {
		el.IdWay = make([]int64, 0)
	}

	if len(el.idWayUnique) == 0 {
		el.idWayUnique = make(map[int64]int64)
	}
	if el.idWayUnique[(*way).Id] != (*way).Id {
		el.idWayUnique[(*way).Id] = (*way).Id
		el.IdWay = append(el.IdWay, (*way).Id)
	}

	//el.Id             =  (*way).Id

	//mapTagLock.Lock()
	//el.Tag            =  (*way).Tag
	//mapTagLock.Unlock()

	el.International = (*way).International
	el.Data = (*way).Data

	for _, loc := range way.Loc {
		el.AddLngLatDegrees(loc[0], loc[1])
	}
	el.Init()
}

// English: Adds all ways that are part of a polygon, then process and generate a single polygon
//
// Português: Adiciona todos os ways que fazem parte de um polígono para depois processar e gerar um polígono único
func (el *PolygonStt) AddWayAsPreProcessingPolygon(way *WayStt) {
	if len(el.TagFromWay) == 0 {
		el.TagFromWay = make(map[string]map[string]string)
	}

	el.TagFromWay[strconv.FormatInt((*way).Id, 10)] = (*way).Tag

	if len(el.IdWay) == 0 {
		el.IdWay = make([]int64, 0)
	}

	if len(el.idWayUnique) == 0 {
		el.idWayUnique = make(map[int64]int64)
	}
	if el.idWayUnique[(*way).Id] != (*way).Id {
		el.idWayUnique[(*way).Id] = (*way).Id
		el.IdWay = append(el.IdWay, (*way).Id)
	}

	var length int = len(way.Loc)
	if length > 0 {
		if way.Loc[0][0] == way.Loc[length-1][0] && way.Loc[0][1] == way.Loc[length-1][1] {
			el.AddWayAsPolygon(way)
			return
		}
	}

	if len(el.tmp) == 0 {
		el.tmp = make([]WayStt, 0)
	}

	el.tmp = append(el.tmp, *way)
}

// English: Transforms the type PointListStt in a polygon
//
// Português: Transforma o tipo PointListStt em um polígono
func (el *PolygonStt) SetPointList(pointList PointListStt) {
	el.PointsList = pointList.List
	el.Initialize = false
}

// English: Transforms the type PointListStt in a polygon and initializes it
//
// Português: Transforma o tipo PointListStt em um polígono e inicializa ele
func (el *PolygonStt) SetPointListAndInit(pointList PointListStt) {
	el.PointsList = pointList.List
	el.Initialize = false
	el.Init()
}

// English: Adds a point in the format latitude and longitude into decimal degrees to the end of the list of points of the polygon
//
// Português: Adiciona um ponto no formato latitude e longitude em graus decimais ao fim da lista de pontos do polígono
func (el *PolygonStt) AddLatLngDegrees(latitudeAFlt, longitudeAFlt float64) {
	if len(el.PointsList) == 0 {
		el.PointsList = make([]PointStt, 0)
	}

	if len(el.Tag) == 0 {
		el.Tag = make(map[string]string)
	}

	var pointLStt PointStt
	pointLStt.SetLatLngDegrees(latitudeAFlt, longitudeAFlt)

	if el.MinimalDistance.Meters != 0 && len(el.PointsList) > 0 {
		var distanceLStt DistanceStt = DistanceBetweenTwoPoints(pointLStt, el.PointsList[len(el.PointsList)-1])
		if distanceLStt.Meters >= el.MinimalDistance.Meters {
			el.PointsList = append(el.PointsList, pointLStt)
		}
		return
	}

	el.PointsList = append(el.PointsList, pointLStt)
}

func (el *PolygonStt) AddPoint(pointAStt *PointStt) {
	if len(el.PointsList) == 0 {
		el.PointsList = make([]PointStt, 0)
	}

	if len(el.Tag) == 0 {
		el.Tag = make(map[string]string)
	}

	if el.MinimalDistance.Meters != 0 && len(el.PointsList) > 0 {
		var distanceLStt DistanceStt = DistanceBetweenTwoPoints(*pointAStt, el.PointsList[len(el.PointsList)-1])
		if distanceLStt.Meters >= el.MinimalDistance.Meters {
			el.PointsList = append(el.PointsList, *pointAStt)
		}
		return
	}

	el.PointsList = append(el.PointsList, *pointAStt)
}

// English: Adds a point in the format latitude and longitude into decimal degrees to the top of the list of points of the polygon
//
// Português: Adiciona um ponto no formato latitude e longitude em graus decimais ao início da lista de pontos do polígono
func (el *PolygonStt) AddLatLngDegreesAtStart(latitudeAFlt, longitudeAFlt float64) {
	if len(el.PointsList) == 0 {
		el.PointsList = make([]PointStt, 0)
	}

	if len(el.Tag) == 0 {
		el.Tag = make(map[string]string)
	}

	var pointLStt PointStt
	pointLStt.SetLatLngDegrees(latitudeAFlt, longitudeAFlt)

	if el.MinimalDistance.Meters != 0 && len(el.PointsList) > 0 {
		var distanceLStt DistanceStt = DistanceBetweenTwoPoints(pointLStt, el.PointsList[0])
		if distanceLStt.Meters >= el.MinimalDistance.Meters {
			el.PointsList = append([]PointStt{pointLStt}, el.PointsList...)
		}
		return
	}

	el.PointsList = append([]PointStt{pointLStt}, el.PointsList...)
}

// English: Adds a point in the format longitude and latitude into decimal degrees to the end of the list of points of the polygon
//
// Português: Adiciona um ponto no formato longitude e latitude em graus decimais ao fim da lista de pontos do polígono
func (el *PolygonStt) AddLngLatDegrees(longitudeAFlt, latitudeAFlt float64) {
	el.AddLatLngDegrees(latitudeAFlt, longitudeAFlt)
}

// English: Adds a point in the format longitude and latitude into decimal degrees to the top of the list of points of the polygon
//
// Português: Adiciona um ponto no formato longitude e latitude em graus decimais ao início da lista de pontos do polígono
func (el *PolygonStt) AddLngLatDegreesAtStart(longitudeAFlt, latitudeAFlt float64) {
	el.AddLatLngDegreesAtStart(latitudeAFlt, longitudeAFlt)
}

// English: Adds a new key on the tag.
//
// Português: Adiciona uma nova chave na tag do polígono.
func (el *PolygonStt) AddTag(keyAStr, valueAStr string) {
	if len(el.Tag) == 0 {
		el.Tag = make(map[string]string)
	}

	el.Tag[keyAStr] = valueAStr
}

// English: Initializes the polygon so that the same can be processed at run-time.
//
// Note that this function must be called on each change in the points of the polygon.
//
// Português: Inicializa o polígono para que o mesmo possa ser processado em tempo de execução.
//
// Note que esta função deve ser chamada a cada alteração nos pontos do polígono.
func (el *PolygonStt) Init() error {
	var distanceList []float64
	var distance float64
	var distanceKey int
	//var distanceAStartBStart float64 = math.MaxFloat64
	//var distanceAStartBEnd float64   = math.MaxFloat64
	//var distanceAEndBStart float64   = math.MaxFloat64
	//var distanceAEndBEnd float64     = math.MaxFloat64
	var k1, k2, lengthStart, lengthTmp int
	//var inverter bool = false
	var pass bool = false
	//var addToEndOfTheSet bool = false
	var pythagorasAStartBStart, pythagorasAStartBEnd, pythagorasAEndBStart, pythagorasAEndBEnd []float64
	var xStart, xEnd, xTmpStart, xTmpEnd float64
	var yStart, yEnd, yTmpStart, yTmpEnd float64

	if len(el.tmp) > 0 {

		for _, loc := range el.tmp[0].Loc {
			el.AddLngLatDegrees(loc[0], loc[1])
		}

		pythagorasAStartBStart = make([]float64, len(el.tmp))
		pythagorasAStartBEnd = make([]float64, len(el.tmp))
		pythagorasAEndBStart = make([]float64, len(el.tmp))
		pythagorasAEndBEnd = make([]float64, len(el.tmp))

		distanceList = make([]float64, len(el.tmp))

		for k2 = 1; k2 != len(el.tmp); k2 += 1 {
			distance = math.MaxFloat64

			for k1 = 0; k1 != len(el.tmp); k1 += 1 {
				distanceList[k1] = math.MaxFloat64
			}

			lengthStart = len(el.PointsList) - 1

			xStart = el.PointsList[0].Loc[0]
			yStart = el.PointsList[0].Loc[1]
			xEnd = el.PointsList[lengthStart].Loc[0]
			yEnd = el.PointsList[lengthStart].Loc[1]

			for k1 = 1; k1 != len(el.tmp); k1 += 1 {

				lengthTmp = len(el.tmp[k1].Loc) - 1

				xTmpStart = el.tmp[k1].Loc[0][0]
				yTmpStart = el.tmp[k1].Loc[0][1]
				xTmpEnd = el.tmp[k1].Loc[lengthTmp][0]
				yTmpEnd = el.tmp[k1].Loc[lengthTmp][1]

				pythagorasAStartBStart[k1] = Pythagoras(xStart, yStart, xTmpStart, yTmpStart)
				pythagorasAStartBEnd[k1] = Pythagoras(xStart, yStart, xTmpEnd, yTmpEnd)
				pythagorasAEndBStart[k1] = Pythagoras(xEnd, yEnd, xTmpStart, yTmpStart)
				pythagorasAEndBEnd[k1] = Pythagoras(xEnd, yEnd, xTmpEnd, yTmpEnd)

				distanceList[k1] = math.Min(distanceList[k1], pythagorasAStartBStart[k1])
				distanceList[k1] = math.Min(distanceList[k1], pythagorasAStartBEnd[k1])
				distanceList[k1] = math.Min(distanceList[k1], pythagorasAEndBStart[k1])
				distanceList[k1] = math.Min(distanceList[k1], pythagorasAEndBEnd[k1])

				if distanceList[k1] < distance {
					distance = distanceList[k1]
					distanceKey = k1
				}
			}

			if pythagorasAStartBStart[distanceKey] < pythagorasAStartBEnd[distanceKey] && pythagorasAStartBStart[distanceKey] < pythagorasAEndBStart[distanceKey] && pythagorasAStartBStart[distanceKey] < pythagorasAEndBEnd[distanceKey] {
				pass = true
				for _, loc := range el.tmp[distanceKey].Loc {
					el.AddLngLatDegreesAtStart(loc[0], loc[1])
				}
			} else if pythagorasAStartBStart[distanceKey] < pythagorasAStartBEnd[distanceKey] && pythagorasAEndBStart[distanceKey] < pythagorasAStartBEnd[distanceKey] && pythagorasAEndBEnd[distanceKey] < pythagorasAStartBEnd[distanceKey] {
				pass = true
				for _, loc := range el.tmp[distanceKey].Loc {
					el.AddLngLatDegrees(loc[0], loc[1])
				}
			} else if pythagorasAStartBStart[distanceKey] < pythagorasAEndBStart[distanceKey] && pythagorasAStartBEnd[distanceKey] < pythagorasAEndBStart[distanceKey] && pythagorasAEndBEnd[distanceKey] < pythagorasAEndBStart[distanceKey] {
				pass = true
				//for _, loc := range el.tmp[k1].Loc {
				for i := len(el.tmp[distanceKey].Loc) - 1; i != -1; i -= 1 {
					el.AddLngLatDegreesAtStart(el.tmp[distanceKey].Loc[i][0], el.tmp[distanceKey].Loc[i][1])
				}
			} else if pythagorasAStartBStart[distanceKey] < pythagorasAEndBEnd[distanceKey] && pythagorasAStartBEnd[distanceKey] < pythagorasAEndBEnd[distanceKey] && pythagorasAEndBStart[distanceKey] < pythagorasAEndBEnd[distanceKey] {
				pass = true
				for i := len(el.tmp[distanceKey].Loc) - 1; i != -1; i -= 1 {
					el.AddLngLatDegrees(el.tmp[distanceKey].Loc[i][0], el.tmp[distanceKey].Loc[i][1])
				}
			} else {
				pass = true
				//for _, loc := range el.tmp[k1].Loc {
				for i := len(el.tmp[distanceKey].Loc) - 1; i != -1; i -= 1 {
					el.AddLngLatDegrees(el.tmp[distanceKey].Loc[i][0], el.tmp[distanceKey].Loc[i][1])
				}
			}

			if pass == true {
				log.Critical("")
			}
		}
	}

	if len(el.PointsList) == 0 {
		return errors.New("polygon has't points")
	}

	el.Initialize = true

	if el.PointsList[0].Loc[0] != el.PointsList[len(el.PointsList)-1].Loc[0] || el.PointsList[0].Loc[1] != el.PointsList[len(el.PointsList)-1].Loc[1] {
		el.PointsList = append(el.PointsList, el.PointsList[0])
	}

	el.ConstantMAFlt = make([]float64, len(el.PointsList))
	el.MultipleMAFlt = make([]float64, len(el.PointsList))
	el.Length = len(el.PointsList)

	lastCornerLUInt := len(el.PointsList) - 1

	for i := 0; i != len(el.PointsList); i += 1 {

		if el.PointsList[lastCornerLUInt].GetLatitudeAsRadians() == el.PointsList[i].GetLatitudeAsRadians() {
			el.ConstantMAFlt[i] = el.PointsList[i].GetLongitudeAsRadians()
			el.MultipleMAFlt[i] = 0
		} else {
			el.ConstantMAFlt[i] = el.PointsList[i].GetLongitudeAsRadians() -
				(el.PointsList[i].GetLatitudeAsRadians()*el.PointsList[lastCornerLUInt].GetLongitudeAsRadians())/
					(el.PointsList[lastCornerLUInt].GetLatitudeAsRadians()-el.PointsList[i].GetLatitudeAsRadians()) +
				(el.PointsList[i].GetLatitudeAsRadians()*el.PointsList[i].GetLongitudeAsRadians())/
					(el.PointsList[lastCornerLUInt].GetLatitudeAsRadians()-el.PointsList[i].GetLatitudeAsRadians())
			el.MultipleMAFlt[i] = (el.PointsList[lastCornerLUInt].GetLongitudeAsRadians() -
				el.PointsList[i].GetLongitudeAsRadians()) /
				(el.PointsList[lastCornerLUInt].GetLatitudeAsRadians() - el.PointsList[i].GetLatitudeAsRadians())
		}

		lastCornerLUInt = i
	}

	el.centroid()
	el.area()

	var pointA = PointStt{}
	var pointB = PointStt{}
	var distanceListLAStt = []DistanceStt{}
	var distanceLStt = DistanceStt{}
	distanceLStt.SetMeters(0.0)

	var k int

	var angleList = []AngleStt{}
	var angle = AngleStt{}
	angle.SetDegrees(0.0)

	distanceListLAStt = make([]DistanceStt, len(el.PointsList))
	distanceListLAStt[0] = distanceLStt

	angleList = make([]AngleStt, len(el.PointsList))

	for keyRefLInt64 := range el.PointsList {
		if keyRefLInt64 != 0 {
			pointA.SetLngLatRadians(el.PointsList[keyRefLInt64-1].Rad[0], el.PointsList[keyRefLInt64-1].Rad[1])
			pointB.SetLngLatRadians(el.PointsList[keyRefLInt64].Rad[0], el.PointsList[keyRefLInt64].Rad[1])

			angleList[keyRefLInt64-1] = DirectionBetweenTwoPoints(pointA, pointB)

			distanceListLAStt[keyRefLInt64] = DistanceBetweenTwoPoints(pointA, pointB)
			distanceLStt.AddMeters(distanceListLAStt[keyRefLInt64].GetMeters())

			k = keyRefLInt64
		}
		angleList[k] = DirectionBetweenTwoPoints(pointA, pointB)
	}

	el.Distance = distanceListLAStt
	el.DistanceTotal = distanceLStt
	el.Angle = angleList
	el.BBox = GetBox(&el.PointsList)
	//el.BBoxBSon = GetBSonBoxInDegrees(&el.PointsList)
	//el.BBoxSearch = [2][2]float64{el.BBox.UpperRight.Loc, el.BBox.BottomLeft.Loc}

	return nil
}

// English: Tests if the point is contained within the polygon.
//
// If the point is above the line of the edge, the same can give a response undetermined because of the lease of the decimals.
//
// Português: Testa se o ponto está contido dentro do polígono.
//
// Se o ponto estiver em cima da linha da borda, o mesmo pode dá uma resposta indeterminada devido ao arrendamento das casas decimais
func (el *PolygonStt) PointInPolygon(pointAStt PointStt) bool {
	if el.Initialize == false {
		el.Initialize = true
		el.Init()
	}

	lastCornerLUInt := len(el.PointsList) - 1
	oddNodesLBoo := false
	tempLBoo := false

	for i := 0; i != len(el.PointsList); i += 1 {
		if el.PointsList[i].GetLatitudeAsRadians() < pointAStt.GetLatitudeAsRadians() &&
			el.PointsList[lastCornerLUInt].GetLatitudeAsRadians() >= pointAStt.GetLatitudeAsRadians() ||
			el.PointsList[lastCornerLUInt].GetLatitudeAsRadians() < pointAStt.GetLatitudeAsRadians() &&
				el.PointsList[i].GetLatitudeAsRadians() >= pointAStt.GetLatitudeAsRadians() {

			tempLBoo = pointAStt.GetLatitudeAsRadians()*el.MultipleMAFlt[i]+el.ConstantMAFlt[i] < pointAStt.GetLongitudeAsRadians()

			// oddNodesLBoo = ( oddNodesLBoo XOR tempLBoo )
			oddNodesLBoo = (oddNodesLBoo || tempLBoo) && !(oddNodesLBoo && tempLBoo)
		}
		lastCornerLUInt = i
	}

	return oddNodesLBoo
}

func (el *PolygonStt) centroid() {
	el.Centroid.Loc = [2]float64{0.0, 0.0}
	el.Centroid.Rad = [2]float64{0.0, 0.0}

	var areaLFlt float64 = 0.0
	var a float64 = 0.0

	var i = 0
	for ; i != len(el.PointsList)-1; i += 1 {
		a = el.PointsList[i].Rad[0]*el.PointsList[i+1].Rad[1] - el.PointsList[i+1].Rad[0]*el.PointsList[i].Rad[1]
		areaLFlt += a
		el.Centroid.Rad[0] += (el.PointsList[i].Rad[0] + el.PointsList[i+1].Rad[0]) * a
		el.Centroid.Rad[1] += (el.PointsList[i].Rad[1] + el.PointsList[i+1].Rad[1]) * a
	}

	a = el.PointsList[i].Rad[0]*el.PointsList[0].Rad[1] - el.PointsList[0].Rad[0]*el.PointsList[i].Rad[1]
	areaLFlt += a
	el.Centroid.Rad[0] += (el.PointsList[i].Rad[0] + el.PointsList[0].Rad[0]) * a
	el.Centroid.Rad[1] += (el.PointsList[i].Rad[1] + el.PointsList[0].Rad[1]) * a

	areaLFlt *= 0.5
	el.Centroid.Rad[0] /= 6.0 * areaLFlt
	el.Centroid.Rad[1] /= 6.0 * areaLFlt

	el.Centroid.Loc[0] = utilMath.RadiansToDegrees(el.Centroid.Rad[0])
	el.Centroid.Loc[1] = utilMath.RadiansToDegrees(el.Centroid.Rad[1])
}

func (el *PolygonStt) area() {
	el.Area = 0.0

	var polygonLStt PolygonStt

	polygonLStt.PointsList = make([]PointStt, len(el.PointsList))

	for i := 0; i != len(el.PointsList); i += 1 {
		earthRadiusLStt := EarthRadius(el.PointsList[i])
		earthRadiusLFlt := earthRadiusLStt.GetKilometers()
		polygonLStt.PointsList[i].SetLatLngRadiansWithoutCheckingFunction(el.PointsList[i].Rad[1]*earthRadiusLFlt, el.PointsList[i].Rad[0]*earthRadiusLFlt)
	}

	var i = 0
	for ; i != len(polygonLStt.PointsList)-1; i += 1 {
		el.Area += polygonLStt.PointsList[i].Rad[0]*polygonLStt.PointsList[i+1].Rad[1] - polygonLStt.PointsList[i+1].Rad[0]*polygonLStt.PointsList[i].Rad[1]
	}

	el.Area += polygonLStt.PointsList[i].Rad[0]*polygonLStt.PointsList[0].Rad[1] - polygonLStt.PointsList[0].Rad[0]*polygonLStt.PointsList[i].Rad[1]
	el.Area *= 0.5
}

// English: Determines the box in which the polygon is contained to be used with the function $box of Mongo DB.
//
// For higher performance of the database, use this function to grab all the points within the box, and then use the function PointInPolygon() to test.
//
// If the point is not in the box, it is not within the polygon.
//
// The answer will be in radians.
//
// Português: Determina a caixa onde o polígono está contido para ser usado com a função $box do MongoDB.
//
// Para maior desempenho do banco, use esta função para pegar todos os pontos dentro da caixa e depois use a função PointInPolygon() para testar.
//
// Caso o ponto não esteja na caixa, ele não está dentro do polígono.
//
// A resposta será em radianos.
func (el *PolygonStt) GetBox() BoxStt { return el.BBox }

// English: Determines the box in which the polygon is contained to be used with the function $box of Mongo DB.
//
// For higher performance of the database, use this function to grab all the points within the box, and then use the function PointInPolygon() to test.
//
// If the point is not in the box, it is not within the polygon.
//
// The answer will be in decimal degrees.
//
// Português: Determina a caixa onde o polígono está contido para ser usado com a função $box do MongoDB.
//
// Para maior desempenho do banco, use esta função para pegar todos os pontos dentro da caixa e depois use a função PointInPolygon() para testar.
//
// Caso o ponto não esteja na caixa, ele não está dentro do polígono.
//
// A resposta será em graus decimais.
//func (el *PolygonStt) GetBSonBoxInDegrees() bson.M { return el.BBoxBSon }

// English: Mounts the geoJSon Feature from polygon and populates the key GeoJSonFeature into the struct
//
// Português: Monta o geoJSon Feature do polígono e popula a chave GeoJSonFeature na struct
func (el *PolygonStt) MakeGeoJSonFeature() string {

	// fixme: fazer
	//if el.Id == 0 {
	//	el.Id = util.AutoId.Get(el.DbCollectionName)
	//}

	var geoJSon GeoJSon = GeoJSon{}
	geoJSon.Init()
	geoJSon.AddGeoMathPolygon(strconv.FormatInt(el.Id, 10), el)
	el.GeoJSonFeature, _ = geoJSon.StringLastFeature()

	return el.GeoJSonFeature
}

// English: Resize a polygon based on the distance between the centroide and the points of construction of the same.
//
// There may be distortions in relation to the polygon and the original polygon.
//
// This function should be improved in the future.
//
// Português: Redimensiona um poligono baseado na distância entre a centroide e os pontos de construção do mesmo.
//
// Pode haver distorções em relação ao polígono original.
//
// Esta função deve ser melhorada em um futuro.
func (el *PolygonStt) Resize(distanceAObj DistanceStt) PolygonStt {
	if el.Initialize == false {
		el.Initialize = true
		el.Init()
	}

	newPolygon := PolygonStt{}
	newPolygon.PointsList = make([]PointStt, len(el.PointsList))

	distance := DistanceStt{}

	direction := AngleStt{}
	point := PointStt{}

	for k, v := range el.PointsList {
		distance = DistanceBetweenTwoPoints(v, el.Centroid)
		distance.AddMeters(distanceAObj.GetMeters())
		direction = DirectionBetweenTwoPoints(v, el.Centroid)
		direction.AddDegrees(180)
		point = DestinationPoint(el.Centroid, distance, direction)
		newPolygon.PointsList[k] = point
	}

	newPolygon.Init()

	return newPolygon
}

// English: Returns the length of the line of the perimeter
//
// Português: Devolve o comprimento da linha de perímetro
func (el *PolygonStt) GetRadius() DistanceStt {
	if el.Initialize == false {
		el.Initialize = true
		el.Init()
	}

	var distanceTmp DistanceStt

	distance := DistanceStt{}
	distance.SetMeters(0.0)

	for _, v := range el.PointsList {
		distanceTmp = DistanceBetweenTwoPoints(v, el.Centroid)
		distance.SetMetersIfGreaterThan(distanceTmp.GetMeters())
	}

	return distance
}

// English: Converts the polygon in a Convex Hull
//
// Special thanks to Valeriy Streltsov for his work in C++
//
// Português: Converte o polígono em um Convex Hull
//
// Agradecimento especial ao Valeriy Streltsov pelo seu trabalho em C++
func (el *PolygonStt) ConvertToConvexHull() {
	var points PointListStt = PointListStt{}
	points.List = el.PointsList
	points = points.ConvexHull()

	el.PointsList = points.List
}

// English: Converts the polygon in a Concave Hull
//
// Special thanks to Valeriy Streltsov for his work in C++
//
// Português: Converte o polígono em um Concave Hull
//
// Agradecimento especial ao Valeriy Streltsov pelo seu trabalho em C++
func (el *PolygonStt) ConvertToConcaveHull(n float64) {
	var points PointListStt = PointListStt{}
	points.List = el.PointsList
	points = points.ConcaveHull(n)

	el.PointsList = points.List
}

// English: Returns a JSon with the information of the polygon.
//
// It is not recommended to backup.
//
// Português: Devolve um JSon com as informações do polígono.
//
// Não é recomendado para backup.
//
// @see ToBSon()
func (el *PolygonStt) ToJSon() ([]byte, error) {
	return bson.MarshalJSON(el)
}

// English: Retorna um io.Reader para escrita em arquivo
//
// Português: Retorna um io.Reader para escrita em arquivo
func (el *PolygonStt) ToReader() io.Reader {
	err, data := el.ToBSon()
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err.Error())
	}

	return bytes.NewReader(data)
}

// English: Loads the polygon based on the data contained in a JSon
//
// Português: Carrega o polígono baseado nos dados contidos em um JSon
func (el *PolygonStt) FromJSon(in []byte) error {
	return bson.UnmarshalJSON(in, el)
}

// English: Writes the polygon in an external file
//
// Português: Escreve o polígono em um arquivo externo
func (el *PolygonStt) ToFile(file io.Writer) error {
	_, err := io.Copy(file, el.ToReader())
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err.Error())
		return err
	}

	return nil
}

// English: Loads the polygon based on the data contained in an external file
//
// Português: Carrega o polígono baseado nos dados contidos em um arquivo externo
func (el *PolygonStt) FromFile(file io.Reader) error {
	var bytesLBty []byte
	var bufferLObj *bytes.Buffer = bytes.NewBuffer(bytesLBty)

	_, err := io.Copy(bufferLObj, file)
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err.Error())
		return err
	}

	err = el.FromBSon(bufferLObj.Bytes())
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err.Error())
		return err
	}

	return nil
}

/*func (el *PolygonStt) FindPointInPolygon(pointQueryAObj bson.M) (error, PointListStt) {
  return el.FindPointInOnePolygon(bson.M{}, pointQueryAObj)
}

func (el *PolygonStt) FindPointInOnePolygon(polygonQueryAObj bson.M, pointQueryAObj bson.M) (error, PointListStt) {
  var err error
  var pointListLStt PointListStt = PointListStt{}

  if el.DbCollectionName == "" {
    el.Prepare()
  }

  pointListLStt = PointListStt{
    DbCollectionName: el.DbCollectionNameForNode,
  }

  var returnPointsLStt PointListStt = PointListStt{}
  returnPointsLStt.List = make([]PointStt, 0)

  if !reflect.DeepEqual(polygonQueryAObj, bson.M{}) {
    err = el.MongoFindOne(polygonQueryAObj)
    if err != nil {
      return err, returnPointsLStt
    }
  }

  if el.Initialize == false {
    el.Init()
  }

  if el.Id == 0 {
    return nil, returnPointsLStt
  }

  if !reflect.DeepEqual(pointQueryAObj, bson.M{}) {
    pointQueryAObj = bson.M{`$and`: []bson.M{{`loc`: el.BBoxBSon}, pointQueryAObj}}
  } else {
    pointQueryAObj = bson.M{`loc`: el.BBoxBSon}
  }

  err = pointListLStt.MongoFind(pointQueryAObj)
  if err != nil {
    log.Criticalf("gOsm.geoMath.geoTypePolygon.Error: ", err.Error())
    return err, returnPointsLStt
  }

  for _, pointToTestLStt := range pointListLStt.List {
    if el.PointInPolygon(pointToTestLStt) == true {
      pointToTestLStt.MongoFindOne(bson.M{"id": pointToTestLStt.Id})
      returnPointsLStt.List = append(returnPointsLStt.List, pointToTestLStt)
    }
  }

  return nil, returnPointsLStt
}*/

func (el *PolygonStt) ToExternalFile(file *os.File, typeId []byte) error {
	var sizeByte = make([]byte, 8)
	var byteBSon []byte
	var err error

	err, byteBSon = el.MakeMD5()
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err)
		return err
	}

	byteBSon, err = zstd.CompressLevel(nil, byteBSon, zstd.DefaultCompression)
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err)
		return err
	}

	binary.LittleEndian.PutUint64(sizeByte, uint64(len(byteBSon)))

	_, err = file.Write(typeId)
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err)
		return err
	}

	_, err = file.Write(sizeByte)
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err)
		return err
	}

	_, err = file.Write(byteBSon)
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err)
	}

	return err
}

func (el *PolygonStt) ToFilePath(filePath string) error {
	var byteBSon []byte
	var err error

	el.Size = 0
	el.Md5 = [16]byte{}
	byteBSon, err = bson.Marshal(el)
	if err != nil {
		log.Criticalf("gOsm.geoMath.tmpPoint.error: %s", err.Error())
	}

	el.Md5 = md5.Sum(byteBSon)
	el.Size = len(byteBSon)

	byteBSon, err = bson.Marshal(el)
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err)
	}

	err = ioutil.WriteFile(filePath, byteBSon, 0644)
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err)
	}

	return err
}

func (el *PolygonStt) FromFilePath(filePath string) error {
	var byteBSon []byte
	var err error

	byteBSon, err = ioutil.ReadFile(filePath)
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err)
	}

	// Transform bson data into point
	err = bson.Unmarshal(byteBSon, el)
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err)
	}

	return err
}

func (el *PolygonStt) CheckMD5() error {
	var err error
	var byteBSon []byte
	var md = el.Md5
	var size = el.Size

	el.Md5 = [16]byte{}
	el.Size = 0
	byteBSon, err = bson.Marshal(el)
	if err != nil {
		log.Criticalf("gOsm.geoMath.tmpPoint.error: %s", err.Error())
	}

	el.Md5 = md5.Sum(byteBSon)
	el.Size = size

	for i := 0; i != 15; i += 1 {
		if el.Md5[i] != md[i] {
			return errors.New("data integrity error")
		}
	}

	return nil
}
