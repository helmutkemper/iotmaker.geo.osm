package iotmaker_geo_osm

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/helmutkemper/gOsm/consts"
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

// point struct based on osm file
type PointStt struct {
	Id int64 `bson:"id"`
	// Array de localização geográfica.
	// [0:x:longitude,1:y:latitude]
	// Este campo deve obrigatoriamente ser um array devido a indexação do MongoDB
	Loc [2]float64 `bson:"loc"`
	Rad [2]float64 `bson:"rad"`

	Visible bool `bson:"visible"`

	// Tags do Open Street Maps
	// As Tags contêm _todo tipo de informação, desde como elas foram importadas, ao nome de um estabelecimento comercial,
	// por exemplo.
	Tag map[string]string `bson:"tag"`

	// Dados do usuário
	// Como o GO é fortemente tipado, eu obtive problemas em estender o struct de forma satisfatória e permitir ao usuário
	// do sistema gravar seus próprios dados, por isto, este campo foi criado. Use-o a vontade.
	Data map[string]string `bson:"data"`

	// Node usado apenas para o parser do arquivo
	GeoJSonFeature string `bson:"geoJSonFeature"`

	Md5  [16]byte `bson:"md5" json:"-"`
	Size int      `bson:"size" json:"-"`

	HasKeyValue bool `bson:"hasKeyValue" json:"-"`
}

// Return type PointStt as []PointStt to be used into gOsm-server project
func (el *PointStt) AsArray() []PointStt {
	var returnLStt []PointStt = make([]PointStt, 1)
	returnLStt[0] = *el

	return returnLStt
}

func (el *PointStt) CopyFrom(pointABStt PointStt) {
	el.Id = pointABStt.Id
	el.Loc = pointABStt.Loc
	el.Rad = pointABStt.Rad
	el.Tag = pointABStt.Tag
	el.Data = pointABStt.Data
	el.Md5 = pointABStt.Md5
	el.Size = pointABStt.Size
}

func (el *PointStt) GetIdAsByte() []byte {
	var ret = make([]byte, 8)
	binary.LittleEndian.PutUint64(ret, uint64(el.Id))

	return ret
}

func (el *PointStt) MakeMD5() (error, []byte) {
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

func (el *PointStt) CheckMD5() error {
	var err error
	var byteBSon []byte
	var md = el.Md5

	el.Size = 0
	el.Md5 = [16]byte{}

	byteBSon, err = bson.Marshal(el)
	if err != nil {
		log.Criticalf("gOsm.geoMath.tmpPoint.error: %s", err.Error())
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

func (el *PointStt) MakeGeoJSonFeature() string {
	var geoJSon GeoJSon = GeoJSon{}
	geoJSon.Init()
	geoJSon.AddGeoMathPoint(strconv.FormatInt(el.Id, 10), el)
	el.GeoJSonFeature, _ = geoJSon.StringLastFeature()

	return el.GeoJSonFeature
}

// Set latitude and longitude as degrees
func (el *PointStt) SetLatLngDegrees(latitudeAFlt, longitudeAFlt float64) error {
	el.Loc = [2]float64{longitudeAFlt, latitudeAFlt}
	el.Rad = [2]float64{utilMath.DegreesToRadians(longitudeAFlt), utilMath.DegreesToRadians(latitudeAFlt)}

	return el.checkBounds()
}

// fixme está estranho...
func (el *PointStt) SetLatLngDecimalDrees(latitudeDegreesAFlt, latitudePrimesAFlt, latitudeSecondsAFlt, longitudeDegreesAFlt, longitudePrimesAFlt, longitudeSecondsAFlt int64) {
	el.Loc = [2]float64{float64(latitudeDegreesAFlt) + float64(latitudePrimesAFlt)/60.0 + float64(latitudeSecondsAFlt)/3600.0, float64(longitudeDegreesAFlt) + float64(longitudePrimesAFlt)/60.0 + float64(longitudeSecondsAFlt)/3600.0}
	el.Rad = [2]float64{utilMath.DegreesToRadians(float64(latitudeDegreesAFlt) + float64(latitudePrimesAFlt)/60.0 + float64(latitudeSecondsAFlt)/3600.0), utilMath.DegreesToRadians(float64(longitudeDegreesAFlt) + float64(longitudePrimesAFlt)/60.0 + float64(longitudeSecondsAFlt)/3600.0)}
}

// fixme está estranho...
func (el *PointStt) SetLngLatDecimalDrees(longitudeDegreesAFlt, longitudePrimesAFlt, longitudeSecondsAFlt, latitudeDegreesAFlt, latitudePrimesAFlt, latitudeSecondsAFlt int64) {
	el.Loc = [2]float64{float64(latitudeDegreesAFlt) + float64(latitudePrimesAFlt)/60.0 + float64(latitudeSecondsAFlt)/3600.0, float64(longitudeDegreesAFlt) + float64(longitudePrimesAFlt)/60.0 + float64(longitudeSecondsAFlt)/3600.0}
	el.Rad = [2]float64{utilMath.DegreesToRadians(float64(latitudeDegreesAFlt) + float64(latitudePrimesAFlt)/60.0 + float64(latitudeSecondsAFlt)/3600.0), utilMath.DegreesToRadians(float64(longitudeDegreesAFlt) + float64(longitudePrimesAFlt)/60.0 + float64(longitudeSecondsAFlt)/3600.0)}
}

// Set longitude and latitude as degrees
func (el *PointStt) SetLngLatDegrees(longitudeAFlt, latitudeAFlt float64) error {
	el.Loc = [2]float64{longitudeAFlt, latitudeAFlt}
	el.Rad = [2]float64{utilMath.DegreesToRadians(longitudeAFlt), utilMath.DegreesToRadians(latitudeAFlt)}

	return el.checkBounds()
}

// Set angle value as degrees
func (el *PointStt) SetXYDegrees(xAFlt, yAFlt float64) error {
	el.Loc = [2]float64{xAFlt, yAFlt}
	el.Rad = [2]float64{utilMath.DegreesToRadians(xAFlt), utilMath.DegreesToRadians(yAFlt)}

	return el.checkBounds()
}

// Set latitude and longitude as radians
func (el *PointStt) SetLatLngRadians(latitudeAFlt, longitudeAFlt float64) error {
	el.Loc = [2]float64{utilMath.RadiansToDegrees(longitudeAFlt), utilMath.RadiansToDegrees(latitudeAFlt)}
	el.Rad = [2]float64{longitudeAFlt, latitudeAFlt}

	return el.checkBounds()
}

func (el *PointStt) SetLatLngRadiansWithoutCheckingFunction(latitudeAFlt, longitudeAFlt float64) {
	el.Loc = [2]float64{utilMath.RadiansToDegrees(longitudeAFlt), utilMath.RadiansToDegrees(latitudeAFlt)}
	el.Rad = [2]float64{longitudeAFlt, latitudeAFlt}
}

// Set longitude and latitude as radians
func (el *PointStt) SetLngLatRadians(longitudeAFlt, latitudeAFlt float64) error {
	el.Loc = [2]float64{utilMath.RadiansToDegrees(longitudeAFlt), utilMath.RadiansToDegrees(latitudeAFlt)}
	el.Rad = [2]float64{longitudeAFlt, latitudeAFlt}

	return el.checkBounds()
}

// Set angle value as radians
func (el *PointStt) SetXYRadians(xAFlt, yAFlt float64) error {
	el.Loc = [2]float64{utilMath.RadiansToDegrees(xAFlt), utilMath.RadiansToDegrees(yAFlt)}
	el.Rad = [2]float64{xAFlt, yAFlt}

	return el.checkBounds()
}

// Get x ( longitude )
func (el *PointStt) GetXAsDegrees() float64 { return el.Loc[0] }

func (el *PointStt) GetXAsRadians() float64 { return el.Rad[0] }

// Get y ( latitude )
func (el *PointStt) GetYDegrees() float64 { return el.Loc[1] }

// Get y ( latitude )
func (el *PointStt) GetYRadians() float64 { return el.Rad[1] }

// Get angle as string
func (el *PointStt) ToRadiansString() string {
	if len(el.Rad) == 0 {
		return fmt.Sprint("(NaN,NaN)")
	}
	return fmt.Sprintf("(%1.5f,%1.5f)%v", el.Rad[0], el.Rad[1], consts.RADIANS)
}

// Get angle as string
func (el *PointStt) ToDegreesString() string {
	if len(el.Loc) == 0 {
		return fmt.Sprint("(NaN,NaN)")
	}
	return fmt.Sprintf("(%1.5f,%1.5f)%v", el.Loc[0], el.Loc[1], consts.DEGREES)
}

// Get latitude and longitude
func (el *PointStt) ToDecimalDegreesString() string {

	dec := math.Abs(el.Loc[0])
	degLng := math.Floor(dec)
	minLng := math.Floor((dec - degLng) * 60.0)
	secLng := (dec - degLng - (minLng / 60.0)) * 3600.0
	if el.Loc[0] < 0 {
		degLng *= -1
	}

	dec = math.Abs(el.Loc[1])
	degLat := math.Floor(dec)
	minLat := math.Floor((dec - degLat) * 60.0)
	secLat := (dec - degLat - (minLat / 60.0)) * 3600.0

	if el.Loc[1] < 0 {
		degLat *= -1
	}

	return fmt.Sprintf("(%v%v%v%v%2.2f%v,%v%v%v%v%2.2f%v)", degLat, consts.DEGREES, minLat, consts.MINUTES, secLat, consts.SECONDS, degLng, consts.DEGREES, minLng, consts.MINUTES, secLng, consts.SECONDS)
}

func (el *PointStt) ToGoogleMapString() string {
	if len(el.Loc) == 0 {
		return fmt.Sprint("(NaN,NaN)")
	}

	return fmt.Sprintf("%1.5f, %1.5f [ Please, copy and past this value on google maps search ]", el.Loc[1], el.Loc[0])
}

func (el *PointStt) ToLeafletMapString() string {
	if len(el.Loc) == 0 {
		return fmt.Sprint("(NaN,NaN)")
	}

	return fmt.Sprintf("[%1.5f, %1.5f],", el.Loc[1], el.Loc[0])
}

// Return y coordinate as latitude
func (el PointStt) GetLatitudeAsDegrees() float64 { return el.Loc[1] }

func (el PointStt) GetLatitudeAsRadians() float64 { return el.Rad[1] }

// Return x coordinate as longitude
func (el PointStt) GetLongitudeAsDegrees() float64 { return el.Loc[0] }

// Return x coordinate as longitude
func (el PointStt) GetLongitudeAsRadians() float64 { return el.Rad[0] }

func (el PointStt) checkBounds() error {
	return nil
	if el.GetLatitudeAsRadians() < consts.MIN_LAT || el.GetLatitudeAsRadians() > consts.MAX_LAT {
		return log.Criticalf("Error: Latitude must be < [math.Pi/2 rad|+90º] and > [-math.Pi/2 rad|-90º]. Value %v\n", el.ToRadiansString())
	}
	if el.GetLongitudeAsRadians() < consts.MIN_LON || el.GetLongitudeAsRadians() > consts.MAX_LON {
		return log.Criticalf("Error: Longitude must be < [math.Pi rad|+180º] and > [-math.Pi rad|-180º]. Value %v\n", el.ToRadiansString())
	}

	return nil
}

func (el PointStt) GetBoundingBox(distanceAStt DistanceStt) BoxStt {
	return BoundingBox(el, distanceAStt)
}

func (el PointStt) GetDestinationPoint(distanceAStt DistanceStt, angleAStt AngleStt) PointStt {
	return DestinationPoint(el, distanceAStt, angleAStt)
}

func (el PointStt) GetDirectionBetweenTwoPoints(pointBAStt PointStt) AngleStt {
	return DirectionBetweenTwoPoints(el, pointBAStt)
}

func (el PointStt) GetDistanceBetweenTwoPoints(pointBAStt PointStt) DistanceStt {
	return DistanceBetweenTwoPoints(el, pointBAStt)
}

func (el *PointStt) Add(pointBAStt PointStt) PointStt {
	var ret PointStt = PointStt{}
	ret.SetLngLatDegrees(el.Loc[0]+pointBAStt.Loc[0], el.Loc[1]+pointBAStt.Loc[1])
	return ret
}

func (el *PointStt) Sub(pointBAStt PointStt) PointStt {
	var ret PointStt = PointStt{}
	ret.SetLngLatDegrees(el.Loc[0]-pointBAStt.Loc[0], el.Loc[1]-pointBAStt.Loc[1])
	return ret
}

func (el *PointStt) Plus(valueAFlt64 float64) PointStt {
	var ret PointStt = PointStt{}
	ret.SetLngLatDegrees(el.Loc[0]*valueAFlt64, el.Loc[1]*valueAFlt64)
	return ret
}

func (el *PointStt) Div(valueAFlt64 float64) PointStt {
	var ret PointStt = PointStt{}
	ret.SetLngLatDegrees(el.Loc[0]/valueAFlt64, el.Loc[1]/valueAFlt64)
	return ret
}

func (el *PointStt) Equality(pointBAStt PointStt) bool {
	return el.Loc[0] == pointBAStt.Loc[0] && el.Loc[1] == pointBAStt.Loc[1]
}

func (el *PointStt) DotProduct(pointBAStt PointStt) float64 {
	return el.Loc[0]*pointBAStt.Loc[0] + el.Loc[1]*pointBAStt.Loc[1]
}

func (el *PointStt) DistanceSquared(pointBAStt PointStt) float64 {
	return (pointBAStt.Loc[0]-el.Loc[0])*(pointBAStt.Loc[0]-el.Loc[0]) + (pointBAStt.Loc[1]-el.Loc[1])*(pointBAStt.Loc[1]-el.Loc[1])
}

func (el *PointStt) Pythagoras(pointBAStt PointStt) float64 {
	return math.Sqrt(el.DistanceSquared(pointBAStt))
}

func (el *PointStt) Distance(pointAAStt, pointBAStt PointStt) float64 {
	var l2 float64 = pointAAStt.DistanceSquared(pointBAStt)
	if l2 == 0.0 {
		return el.Pythagoras(pointAAStt) // v == w case
	}

	// Consider the line extending the segment, parameterized as v + t (w - v)
	// We find projection of point p onto the line.
	// It falls where t = [(p-v) . (w-v)] / |w-v|^2
	var pA PointStt = el.Sub(pointAAStt)
	var pB PointStt = pointBAStt.Sub(pointAAStt)
	var t float64 = pA.DotProduct(pB) / l2
	if t < 0.0 {
		return el.Pythagoras(pointAAStt)
	} else if t > 1.0 {
		return el.Pythagoras(pointBAStt)
	}
	var pC PointStt = pointBAStt.Sub(pointAAStt)
	pC = pC.Plus(t)
	pC = pointAAStt.Add(pC)

	return el.Pythagoras(pC)
}

func (el *PointStt) DecisionDistance(pointsAAStt []PointStt) float64 {
	var i int
	var curDistance float64
	var dst float64 = el.Pythagoras(pointsAAStt[0])
	for i = 1; i < len(pointsAAStt); i += 1 {
		curDistance = el.Pythagoras(pointsAAStt[i])
		if curDistance < dst {
			dst = curDistance
		}
	}

	return dst
}

func (el *PointStt) IsContainedInTheList(pointsAAStt []PointStt) bool {
	for _, point := range pointsAAStt {
		if el.Equality(point) {
			return true
		}
	}

	return false
}

func (el *PointStt) ToExternalFile(file *os.File, typeId []byte) error {
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

func (el *PointStt) ToFilePath(filePath string) error {
	var byteBSon []byte
	var err error

	err, byteBSon = el.MakeMD5()
	if err != nil {
		log.Criticalf("gOsm.geoMath.tmpPoint.error: %s", err.Error())
	}

	err = ioutil.WriteFile(filePath, byteBSon, 0644)
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err)
	}

	return err
}

func (el *PointStt) FromFilePath(filePath string) error {
	var byteBSon []byte
	var err error

	byteBSon, err = ioutil.ReadFile(filePath)
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err)
	}

	el.FromBSon(byteBSon)

	return err
}

func (el *PointStt) RemoveFilePath(filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		return errors.New("%v is a dir, not a file")
	}

	return os.Remove(filePath)
}

func (el *PointStt) ToBSon() (error, []byte) {
	return el.MakeMD5()
}

func (el *PointStt) ToJSon() (error, []byte) {
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

	byteBSon, err = bson.MarshalJSON(el)

	return err, byteBSon
}

func (el *PointStt) ToReader() io.Reader {
	err, data := el.ToBSon()
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err)
	}

	return bytes.NewReader(data)
}

func (el *PointStt) FromBSon(in []byte) error {
	return bson.Unmarshal(in, el)
}

func (el *PointStt) FromJSon(in []byte) error {
	return bson.UnmarshalJSON(in, el)
}

func (el *PointStt) ToFile(file io.Writer) error {
	_, err := io.Copy(file, el.ToReader())
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err.Error())
		return err
	}

	return nil
}

func (el *PointStt) FromFile(file io.Reader) error {
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
