package iotmaker_geo_osm

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/helmutkemper/gOsm/consts"
	"github.com/helmutkemper/mgo/bson"
	"github.com/helmutkemper/zstd"
	"io"
	"io/ioutil"
	"math"
	"os"
	"strconv"
)

type WayStt struct {
	Id            int64             `bson:"id"`
	Visible       bool              `bson:"visible"`
	Tag           map[string]string `bson:"tag"`
	International map[string]string `bson:"international"`
	Loc           [][2]float64      `bson:"loc"`
	LocFirst      [2]float64        `bson:"locFirst"`
	LocLast       [2]float64        `bson:"locLast"`
	Rad           [][2]float64      `bson:"rad"`
	Distance      []DistanceStt     `bson:"distance"`
	DistanceTotal DistanceStt       `bson:"distanceTotal"`
	Angle         []AngleStt        `bson:"angle"`

	Data map[string]string `bson:"data"`
	// en: boundary box in degrees
	// pt: caixa de perímetro em graus decimais
	BBox           BoxStt `bson:"bbox"`
	GeoJSonFeature string `bson:"geoJSonFeature"`

	SurroundingPreset []float64 `bson:"SurroundingPreset" json:"-"`

	Md5  [16]byte `bson:"md5" json:"-"`
	Size int      `bson:"size" json:"-"`

	HasKeyValue bool `bson:"hasKeyValue" json:"-"`
}

func (el *WayStt) GetIdAsByte() []byte {
	var ret = make([]byte, 8)
	binary.LittleEndian.PutUint64(ret, uint64(el.Id))

	return ret
}

func (el *WayStt) MakeMD5() (error, []byte) {
	var err error
	var byteBSon []byte

	el.Size = 0
	el.Md5 = [16]byte{}
	byteBSon, err = bson.Marshal(el)
	if err != nil {
		return err, []byte{}
	}

	el.Md5 = md5.Sum(byteBSon)
	el.Size = len(byteBSon)

	byteBSon, err = bson.Marshal(el)
	if err != nil {
		return err, []byte{}
	}

	return nil, byteBSon
}

func (el *WayStt) CheckMD5() error {
	var err error
	var byteBSon []byte
	var md = el.Md5

	el.Size = 0
	el.Md5 = [16]byte{}

	byteBSon, err = bson.Marshal(el)
	if err != nil {
		return err
	}

	el.Md5 = md5.Sum(byteBSon)
	el.Size = len(byteBSon)

	for i := 0; i != 15; i += 1 {
		if el.Md5[i] != md[i] {
			return errors.New("data integrity error")
		}
	}

	return nil
}

func (el *WayStt) ToBSon() (error, []byte) {
	return el.MakeMD5()
}

func (el *WayStt) FromBSon(byteBSon []byte) error {
	var err error

	err = bson.Unmarshal(byteBSon, el)
	return err
}

//fixme código estranho no ângulo
func (el *WayStt) Init() error {
	var err error

	var longitudeMaxLFlt float64
	var longitudeMinLFlt float64
	var latitudeMaxLFlt float64
	var latitudeMinLFlt float64

	longitudeMaxLFlt = -999.9
	longitudeMinLFlt = 999.9
	latitudeMaxLFlt = -999.9
	latitudeMinLFlt = 999.9

	var pointA = PointStt{}
	var pointB = PointStt{}
	var distanceList = make([]DistanceStt, 0)
	var distance = DistanceStt{}
	distance.SetMeters(0.0)

	var angleList = make([]AngleStt, 0)
	var angle = AngleStt{}
	angle.SetDegrees(0.0)

	distanceList = make([]DistanceStt, len(el.Rad))
	distanceList[0] = distance

	angleList = make([]AngleStt, len(el.Rad))

	for keyRefLInt64 := range el.Rad {
		if keyRefLInt64 != 0 {
			err = pointA.SetLngLatRadians(el.Rad[keyRefLInt64-1][0], el.Rad[keyRefLInt64-1][1])
			if err != nil {
				return err
			}

			err = pointB.SetLngLatRadians(el.Rad[keyRefLInt64][0], el.Rad[keyRefLInt64][1])
			if err != nil {
				return err
			}

			angleList[keyRefLInt64-1] = DirectionBetweenTwoPoints(pointA, pointB)

			distanceList[keyRefLInt64] = DistanceBetweenTwoPoints(pointA, pointB)
			distance.AddMeters(distanceList[keyRefLInt64].GetMeters())
		}

		longitudeMaxLFlt = math.Max(longitudeMaxLFlt, el.Loc[keyRefLInt64][0])
		longitudeMinLFlt = math.Min(longitudeMinLFlt, el.Loc[keyRefLInt64][0])
		latitudeMaxLFlt = math.Max(latitudeMaxLFlt, el.Loc[keyRefLInt64][1])
		latitudeMinLFlt = math.Min(latitudeMinLFlt, el.Loc[keyRefLInt64][1])
	}

	var indexMaxLInt = len(el.Rad) - 1
	if indexMaxLInt > 1 {
		angleList[indexMaxLInt] = angleList[indexMaxLInt-1]
	}

	if len(el.Loc) != 0 {
		el.LocFirst = el.Loc[0]
		el.LocLast = el.Loc[len(el.Loc)-1]
	}

	el.Distance = distanceList
	el.DistanceTotal = distance
	el.Angle = angleList
	el.BBox = GetBoxFlt(&el.Loc)

	return nil
}

func (el *WayStt) IsPolygon() bool {
	var length int = len(el.Loc) - 1
	if length < 2 {
		return false
	}

	if el.Loc[0][0] == el.Loc[length][0] && el.Loc[0][1] == el.Loc[length][1] {
		return true
	}

	return false
}

func (el *WayStt) AddLatLngDegreesAtStart(latitudeAFlt, longitudeAFlt float64) error {
	if len(el.Loc) == 0 {
		el.Loc = make([][2]float64, 0)
		el.Rad = make([][2]float64, 0)
	}

	el.Loc = append([][2]float64{{longitudeAFlt, latitudeAFlt}}, el.Loc...)
	el.Rad = append([][2]float64{{DegreesToRadians(longitudeAFlt), DegreesToRadians(latitudeAFlt)}}, el.Rad...)

	return el.checkBounds()
}

func (el *WayStt) AddLngLatDegreesAtStart(longitudeAFlt, latitudeAFlt float64) error {
	if len(el.Loc) == 0 {
		el.Loc = make([][2]float64, 0)
		el.Rad = make([][2]float64, 0)
	}

	el.Loc = append([][2]float64{{longitudeAFlt, latitudeAFlt}}, el.Loc...)
	el.Rad = append([][2]float64{{DegreesToRadians(longitudeAFlt), DegreesToRadians(latitudeAFlt)}}, el.Rad...)

	return el.checkBounds()
}

func (el *WayStt) AddLatLngDegrees(latitudeAFlt, longitudeAFlt float64) error {
	if len(el.Loc) == 0 {
		el.Loc = make([][2]float64, 0)
		el.Rad = make([][2]float64, 0)
	}

	el.Loc = append(el.Loc, [2]float64{longitudeAFlt, latitudeAFlt})
	el.Rad = append(el.Rad, [2]float64{DegreesToRadians(longitudeAFlt), DegreesToRadians(latitudeAFlt)})

	return el.checkBounds()
}

func (el *WayStt) AddLngLatDegrees(longitudeAFlt, latitudeAFlt float64) error {
	if len(el.Loc) == 0 {
		el.Loc = make([][2]float64, 0)
		el.Rad = make([][2]float64, 0)
	}

	el.Loc = append(el.Loc, [2]float64{longitudeAFlt, latitudeAFlt})
	el.Rad = append(el.Rad, [2]float64{DegreesToRadians(longitudeAFlt), DegreesToRadians(latitudeAFlt)})

	return el.checkBounds()
}

func (el *WayStt) AddXYDegrees(xAFlt, yAFlt float64) error {
	if len(el.Loc) == 0 {
		el.Loc = make([][2]float64, 0)
		el.Rad = make([][2]float64, 0)
	}

	el.Loc = append(el.Loc, [2]float64{xAFlt, yAFlt})
	el.Rad = append(el.Rad, [2]float64{DegreesToRadians(xAFlt), DegreesToRadians(yAFlt)})

	return el.checkBounds()
}

func (el *WayStt) AddLatLngRadians(latitudeAFlt, longitudeAFlt float64) error {
	if len(el.Loc) == 0 {
		el.Loc = make([][2]float64, 0)
		el.Rad = make([][2]float64, 0)
	}

	el.Loc = append(el.Loc, [2]float64{RadiansToDegrees(longitudeAFlt), RadiansToDegrees(latitudeAFlt)})
	el.Rad = append(el.Rad, [2]float64{longitudeAFlt, latitudeAFlt})

	return el.checkBounds()
}

func (el *WayStt) AddLngLatRadians(longitudeAFlt, latitudeAFlt float64) error {
	if len(el.Loc) == 0 {
		el.Loc = make([][2]float64, 0)
		el.Rad = make([][2]float64, 0)
	}

	el.Loc = append(el.Loc, [2]float64{RadiansToDegrees(longitudeAFlt), RadiansToDegrees(latitudeAFlt)})
	el.Rad = append(el.Rad, [2]float64{longitudeAFlt, latitudeAFlt})

	return el.checkBounds()
}

func (el *WayStt) AddXYRadians(xAFlt, yAFlt float64) error {
	if len(el.Loc) == 0 {
		el.Loc = make([][2]float64, 0)
		el.Rad = make([][2]float64, 0)
	}

	el.Loc = append(el.Loc, [2]float64{RadiansToDegrees(xAFlt), RadiansToDegrees(yAFlt)})
	el.Rad = append(el.Rad, [2]float64{xAFlt, yAFlt})

	return el.checkBounds()
}

func (el *WayStt) AddPoint(pointAStt *PointStt) error {
	if len(el.Loc) == 0 {
		el.Loc = make([][2]float64, 0)
		el.Rad = make([][2]float64, 0)
	}

	el.Loc = append([][2]float64{{pointAStt.Loc[0], pointAStt.Loc[1]}}, el.Loc...)
	el.Rad = append([][2]float64{{pointAStt.Rad[0], pointAStt.Rad[1]}}, el.Rad...)

	return el.checkBounds()
}

func (el *WayStt) SetId(idAUI64 int64) {
	el.Id = idAUI64
}

// en: Adds a new key on the tag.
//
// pt: Adiciona uma nova chave na tag.
func (el *WayStt) AddTag(keyAStr, valueAStr string) {
	if len(el.Tag) == 0 {
		el.Tag = make(map[string]string)
	}

	el.Tag[keyAStr] = valueAStr
}

func (el *WayStt) checkBounds() error {
	if el.Rad[len(el.Rad)-1][1] < consts.MIN_LAT || el.Rad[len(el.Rad)-1][1] > consts.MAX_LAT {
		return fmt.Errorf("Error: Latitude must be < [math.Pi/2 rad|+90º] and > [-math.Pi/2 rad|-90º]. Value (%1.5f,%1.5f)%v\n", el.Rad[len(el.Rad)-1][0], el.Rad[len(el.Rad)-1][1], consts.RADIANS)
	}
	if el.Rad[len(el.Rad)-1][0] < consts.MIN_LON || el.Rad[len(el.Rad)-1][0] > consts.MAX_LON {
		return fmt.Errorf("Error: Longitude must be < [math.Pi rad|+180º] and > [-math.Pi rad|-180º]. Value (%1.5f,%1.5f)%v\n", el.Rad[len(el.Rad)-1][0], el.Rad[len(el.Rad)-1][1], consts.RADIANS)
	}

	return nil
}

func (el *WayStt) MakeGeoJSonFeature() string {
	var geoJSon GeoJSon = GeoJSon{}
	geoJSon.Init()
	geoJSon.AddGeoMathWay(strconv.FormatInt(el.Id, 10), el)
	el.GeoJSonFeature, _ = geoJSon.StringLastFeature()

	return el.GeoJSonFeature
}

func (el *WayStt) MakePolygonSurroundings(distanceAStt, minimalDistanceAStt DistanceStt) (error, PolygonStt) {
	if len(el.Loc) < 3 {
		return errors.New("the way must have a minimum of three points"), PolygonStt{}
	}

	el.Init()

	var angleLStt AngleStt = el.Angle[0]
	angleLStt.AddDegrees(60)

	var polygonLStt PolygonStt = PolygonStt{}
	polygonLStt.SetMinimalDistance(minimalDistanceAStt)

	var pointALStt, pointBLStt PointStt

	pointALStt.SetLngLatDegrees(el.Loc[0][0], el.Loc[0][1])

	for i := 0; i != 7; i += 1 {
		angleLStt.AddDegrees(30)
		pointBLStt = DestinationPoint(pointALStt, distanceAStt, angleLStt)
		polygonLStt.AddPoint(&pointBLStt)
	}

	for i := 0; i != len(el.Loc)-1; i += 1 {
		pointALStt.SetLngLatDegrees(el.Loc[i][0], el.Loc[i][1])

		angleLStt = el.Angle[i]
		angleLStt.AddDegrees(-90)

		pointBLStt = DestinationPoint(pointALStt, distanceAStt, angleLStt)
		polygonLStt.AddPoint(&pointBLStt)
	}

	angleLStt = el.Angle[len(el.Loc)-1]
	angleLStt.AddDegrees(-120)

	for i := 0; i != 7; i += 1 {
		angleLStt.AddDegrees(30)
		pointALStt.SetLngLatDegrees(el.Loc[len(el.Loc)-1][0], el.Loc[len(el.Loc)-1][1])
		pointBLStt = DestinationPoint(pointALStt, distanceAStt, angleLStt)
		polygonLStt.AddPoint(&pointBLStt)
	}

	for i := len(el.Loc) - 1; i != -1; i -= 1 {
		pointALStt.SetLngLatDegrees(el.Loc[i][0], el.Loc[i][1])

		angleLStt = el.Angle[i]
		angleLStt.AddDegrees(90)

		pointBLStt = DestinationPoint(pointALStt, distanceAStt, angleLStt)
		polygonLStt.AddPoint(&pointBLStt)
	}

	pointALStt.SetLngLatDegrees(el.Loc[0][0], el.Loc[0][1])

	angleLStt = el.Angle[0]
	angleLStt.AddDegrees(90)

	pointBLStt = DestinationPoint(pointALStt, distanceAStt, angleLStt)
	polygonLStt.AddPoint(&pointBLStt)

	polygonLStt.AddWayDataAsPolygonData(el)
	polygonLStt.Id = el.Id
	polygonLStt.Init()

	return nil, polygonLStt
}

func (el *WayStt) MakePolygonSurroundingsRight(distanceAStt, minimalDistanceAStt DistanceStt) (error, PolygonStt) {
	if len(el.Loc) < 3 {
		return errors.New("the way must have a minimum of three points"), PolygonStt{}
	}

	el.Init()

	var angleLStt AngleStt = el.Angle[0]
	angleLStt.AddDegrees(-30)

	var polygonLStt PolygonStt = PolygonStt{}
	polygonLStt.SetMinimalDistance(minimalDistanceAStt)

	var pointALStt, pointBLStt PointStt

	angleLStt = el.Angle[len(el.Loc)-1]
	angleLStt.AddDegrees(-120)

	for i := 0; i != len(el.Loc)-1; i += 1 {
		pointALStt.SetLngLatDegrees(el.Loc[i][0], el.Loc[i][1])

		angleLStt = el.Angle[i]
		angleLStt.AddDegrees(-90)

		pointBLStt = DestinationPoint(pointALStt, distanceAStt, angleLStt)
		polygonLStt.AddPoint(&pointBLStt)
	}

	for i := 0; i != 3; i += 1 {
		angleLStt.AddDegrees(30)
		pointALStt.SetLngLatDegrees(el.Loc[len(el.Loc)-1][0], el.Loc[len(el.Loc)-1][1])
		pointBLStt = DestinationPoint(pointALStt, distanceAStt, angleLStt)
		polygonLStt.AddPoint(&pointBLStt)
	}

	for i := len(el.Loc) - 1; i != -1; i -= 1 {
		pointALStt.SetLngLatDegrees(el.Loc[i][0], el.Loc[i][1])
		polygonLStt.AddPoint(&pointALStt)
	}

	angleLStt = el.Angle[0]
	angleLStt.AddDegrees(150)
	for i := 0; i != 4; i += 1 {
		angleLStt.AddDegrees(30)
		pointALStt.SetLngLatDegrees(el.Loc[0][0], el.Loc[0][1])
		pointBLStt = DestinationPoint(pointALStt, distanceAStt, angleLStt)
		polygonLStt.AddPoint(&pointBLStt)
	}

	polygonLStt.AddWayDataAsPolygonData(el)
	polygonLStt.Id = el.Id
	polygonLStt.Init()

	return nil, polygonLStt
}

func (el *WayStt) MakePolygonSurroundingsLeft(distanceAStt, minimalDistanceAStt DistanceStt) (error, PolygonStt) {
	if len(el.Loc) < 3 {
		return errors.New("the way must have a minimum of three points"), PolygonStt{}
	}

	el.Init()

	var angleLStt AngleStt = el.Angle[0]
	angleLStt.AddDegrees(60)

	var polygonLStt PolygonStt = PolygonStt{}
	polygonLStt.SetMinimalDistance(minimalDistanceAStt)

	var pointALStt, pointBLStt PointStt

	pointALStt.SetLngLatDegrees(el.Loc[0][0], el.Loc[0][1])

	for i := 0; i != 4; i += 1 {
		angleLStt.AddDegrees(30)
		pointBLStt = DestinationPoint(pointALStt, distanceAStt, angleLStt)
		polygonLStt.AddPoint(&pointBLStt)
	}

	for i := 0; i != len(el.Loc)-1; i += 1 {
		pointALStt.SetLngLatDegrees(el.Loc[i][0], el.Loc[i][1])
		polygonLStt.AddPoint(&pointALStt)
	}

	angleLStt = el.Angle[len(el.Loc)-1]
	angleLStt.AddDegrees(-30)

	for i := 0; i != 4; i += 1 {
		angleLStt.AddDegrees(30)
		pointALStt.SetLngLatDegrees(el.Loc[len(el.Loc)-1][0], el.Loc[len(el.Loc)-1][1])
		pointBLStt = DestinationPoint(pointALStt, distanceAStt, angleLStt)
		polygonLStt.AddPoint(&pointBLStt)
	}

	for i := len(el.Loc) - 1; i != -1; i -= 1 {
		pointALStt.SetLngLatDegrees(el.Loc[i][0], el.Loc[i][1])

		angleLStt = el.Angle[i]
		angleLStt.AddDegrees(90)

		pointBLStt = DestinationPoint(pointALStt, distanceAStt, angleLStt)
		polygonLStt.AddPoint(&pointBLStt)
	}

	polygonLStt.AddWayDataAsPolygonData(el)
	polygonLStt.Id = el.Id
	polygonLStt.Init()

	return nil, polygonLStt
}

func (el *WayStt) ToJSon() ([]byte, error) {
	return bson.MarshalJSON(el)
}

func (el *WayStt) ToReader() (error, io.Reader) {
	err, data := el.ToBSon()
	if err != nil {
		return err, nil
	}

	return nil, bytes.NewReader(data)
}

func (el *WayStt) FromJSon(in []byte) error {
	return bson.UnmarshalJSON(in, el)
}

func (el *WayStt) ToFile(file io.Writer) error {
	err, reader := el.ToReader()
	if err != nil {
		return err
	}

	_, err = io.Copy(file, reader)
	if err != nil {
		return err
	}

	return nil
}

func (el *WayStt) FromFile(file io.Reader) error {
	var bytesLBty []byte
	var bufferLObj *bytes.Buffer = bytes.NewBuffer(bytesLBty)

	_, err := io.Copy(bufferLObj, file)
	if err != nil {
		return err
	}

	err = el.FromBSon(bufferLObj.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (el *WayStt) ToExternalFile(file *os.File, typeId []byte) error {
	var sizeByte = make([]byte, 8)
	var byteBSon []byte
	var err error

	err, byteBSon = el.MakeMD5()
	if err != nil {
		return err
	}

	byteBSon, err = zstd.CompressLevel(nil, byteBSon, zstd.DefaultCompression)
	if err != nil {
		return err
	}

	binary.LittleEndian.PutUint64(sizeByte, uint64(len(byteBSon)))

	_, err = file.Write(typeId)
	if err != nil {
		return err
	}

	_, err = file.Write(sizeByte)
	if err != nil {
		return err
	}

	_, err = file.Write(byteBSon)
	return err
}

func (el *WayStt) ToFilePath(filePath string) error {
	var byteBSon []byte
	var err error

	byteBSon, err = bson.Marshal(el)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filePath, byteBSon, 0644)
	return err
}

func (el *WayStt) FromFilePath(filePath string) error {
	var byteBSon []byte
	var err error

	byteBSon, err = ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Transform bson data into point
	err = bson.Unmarshal(byteBSon, el)
	return err
}
