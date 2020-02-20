package iotmaker_geo_osm

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"github.com/helmutkemper/mgo/bson"
	log "github.com/helmutkemper/seelog"
	"github.com/helmutkemper/zstd"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

type PolygonListStt struct {
	// id do open street maps
	Id int64 `bson:"id"`
	// Versão dentro do Open Street Maps
	Version int64 `bson:"version"`
	// TimeStamp dentro do Open Street Maps
	TimeStamp time.Time `bson:"timeStamp"`
	// ChangeSet dentro do Open Street Maps
	ChangeSet int64 `bson:"changeSet"`

	Visible bool `bson:"visible"`

	// User Id dentro do Open Street Maps
	UId int64 `bson:"userId"`
	// User Name dentro do Open Street Maps
	User string `bson:"-"`
	// Tags do Open Street Maps
	// As Tags contêm _todo tipo de informação, desde como elas foram importadas, ao nome de um estabelecimento comercial,
	// por exemplo.
	Tag           map[string]string            `bson:"tag"`
	TagFromWay    map[string]map[string]string `bson:"tagFromWay"`
	International map[string]string            `bson:"international"`
	Initialized   bool                         `bson:"inicialize"`
	// Dados do usuário
	// Como o GO é fortemente tipado, eu obtive problemas em estender o struct de forma satisfatória e permitir ao usuário
	// do sistema gravar seus próprios dados, por isto, este campo foi criado. Use-o a vontade.
	Data map[string]string `bson:"data"`
	Role string            `bson:"role"`

	idRelationUnique map[int64]int64 `bson:"-"`
	idPolygonUnique  map[int64]int64 `bson:"-"`
	idWayUnique      map[int64]int64 `bson:"-"`

	IdRelation []int64      `bson:"idRelation"`
	IdPolygon  []int64      `bson:"idPolygon"`
	IdWay      []int64      `bson:"idWay"`
	List       []PolygonStt `bson:"list"`
	// en: boundary box in degrees
	// pt: caixa de perímetro em graus decimais
	BBox BoxStt `bson:"bbox"`
	// en: boundary box in BSon to MongoDB
	// pt: caixa de perímetro em BSon para o MongoDB
	BBoxBSon       bson.M `bson:"bboxBSon"`
	GeoJSon        string `bson:"geoJSon"`
	GeoJSonFeature string `bson:"geoJSonFeature"`

	Md5  [16]byte `bson:"md5" json:"-"`
	Size int      `bson:"size" json:"-"`

	HasKeyValue bool `bson:"hasKeyValue" json:"-"`
}

func (el *PolygonListStt) GetIdAsByte() []byte {
	var ret = make([]byte, 8)
	binary.LittleEndian.PutUint64(ret, uint64(el.Id))

	return ret
}

func (el *PolygonListStt) MakeMD5() (error, []byte) {
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

func (el *PolygonListStt) CheckMD5() error {
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

func (el *PolygonListStt) ToBSon() (error, []byte) {
	return el.MakeMD5()
}

func (el *PolygonListStt) FromBSon(byteBSon []byte) error {
	var err error

	err = bson.Unmarshal(byteBSon, el)
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err)
	}

	return err
}

// en: Copies the data of the relation in the polygon.
//
// pt: Copia os dados de uma relação no polígono.
//
// @see dataOfOsm in blog.osm.io
func (el *PolygonListStt) AddRelationDataAsPolygonData(relation *RelationStt) {
	if len(el.IdRelation) == 0 {
		el.IdRelation = make([]int64, 0)
	}
	if len(el.idRelationUnique) == 0 {
		el.idRelationUnique = make(map[int64]int64)
	}
	if el.idRelationUnique[(*relation).Id] != (*relation).Id {
		el.idRelationUnique[(*relation).Id] = (*relation).Id
		el.IdRelation = append(el.IdRelation, (*relation).Id)
	}

	el.Version = (*relation).Version
	el.TimeStamp = (*relation).TimeStamp
	el.ChangeSet = (*relation).ChangeSet
	el.Visible = (*relation).Visible
	el.UId = (*relation).UId
	el.User = (*relation).User
	el.Tag = (*relation).Tag
	el.International = (*relation).International
}

func (el *PolygonListStt) AddWayAsPolygon(way *WayStt) {
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

	var polygon PolygonStt = PolygonStt{}
	polygon.AddWayAsPolygon(way)
	polygon.Init()

	el.AddPolygon(&polygon)
}

func (el *PolygonListStt) AddPolygon(polygon *PolygonStt) {
	if len(el.TagFromWay) == 0 {
		el.TagFromWay = make(map[string]map[string]string)
	}
	el.TagFromWay[strconv.FormatInt(polygon.Id, 10)] = polygon.Tag

	if len(el.IdPolygon) == 0 {
		el.IdPolygon = make([]int64, 0)
	}
	if len(el.idPolygonUnique) == 0 {
		el.idPolygonUnique = make(map[int64]int64)
	}
	if el.idPolygonUnique[(*polygon).Id] != (*polygon).Id {
		el.idPolygonUnique[(*polygon).Id] = (*polygon).Id
		el.IdPolygon = append(el.IdPolygon, polygon.Id)
	}

	if len(el.List) == 0 {
		el.List = make([]PolygonStt, 0)
	}

	el.List = append(el.List, *polygon)
}

/*func ( el *PolygonListStt ) FindFromPolygon( queryAObj bson.M ) error {
  err := el.DbStt.TestConnection()
  if err != nil{
    log.Criticalf( "gOsm.geoMath.geoTypePolygon.error: %s", err )
    return err
  }

  err = el.DbStt.Find( el.DbCollectionName, &el.List, queryAObj )
  if err != nil{
    log.Criticalf( "gOsm.geoMath.geoTypePolygon.error: %s", err )
    return err
  }

  var polygonTmp = PolygonStt{}

  for k, polygon := range el.List {
    polygonTmp = PolygonStt{
      DbCollectionName: el.DbCollectionName,
    }

    err = polygonTmp.FromGrid( polygon.Id )
    if err != nil{
      log.Criticalf( "gOsm.geoMath.geoTypePolygon.error: %s", err )
      return err
    }

    el.List[ k ] = polygonTmp
  }

  return err
}*/

/*func ( el *PolygonListStt ) FindPolygonByLngLatDegrees( lng, lat float64 ) error {
  var err error
  var polygonList = PolygonListStt{}
  var point = PointStt{}
  point.SetLngLatDegrees( lng, lat )

  if el.DbCollectionName == "" {
    el.Prepare()
  }

  err = el.DbStt.Find( el.DbCollectionNameForPolygon, &el.List, bson.M{ "bBoxSearch.1.0": bson.M{ "$gte": lng }, "bBoxSearch.1.1": bson.M{ "$lte": lat }, "bBoxSearch.0.0": bson.M{ "$lte": lng }, "bBoxSearch.0.1": bson.M{ "$gte": lat } } )
  if err != nil{
    log.Criticalf( "gOsm.geoMath.geoTypePolygon.error: %s", err.Error() )
    return err
  }

  for _, polygon := range el.List{
    if polygon.HasKeyValue == true {
      err = polygon.DbKeyValueFind(polygon.Id)
      if err != nil {
        log.Criticalf("gOsm.geoMath.el.error: %s", err.Error())
        return err
      }
    }

    if polygon.PointInPolygon( point ) {
      polygonList.AddPolygon( &polygon )
    }
  }

  el.List = make( []PolygonStt, 0 )
  for _, polygon := range polygonList.List{
    el.List = append( el.List, polygon )
  }

  return nil
}*/

func (el *PolygonListStt) MakeGeoJSonAllFeatures() string {

	//if el.Id == 0 {
	//  el.Id = util.AutoId.Get( consts.DB_OSM_FILE_POLYGONS_COLLECTIONS )
	//}

	var geoJSon GeoJSon = GeoJSon{}
	geoJSon.Init()
	geoJSon.AddGeoMathPolygonList(strconv.FormatInt(el.Id, 10), el)
	el.GeoJSonFeature, _ = geoJSon.StringAllFeatures()

	return el.GeoJSonFeature
}

func (el *PolygonListStt) MakeGeoJSon() string {
	//if el.Id == 0 {
	//  el.Id = util.AutoId.Get( consts.DB_OSM_FILE_POLYGONS_COLLECTIONS )
	//}

	var geoJSon GeoJSon = GeoJSon{}
	geoJSon.Init()
	geoJSon.AddGeoMathPolygonList(strconv.FormatInt(el.Id, 10), el)
	el.GeoJSon, _ = geoJSon.String()

	return el.GeoJSon
}

func (el *PolygonListStt) MakeGeoJSonFeature() string {
	//if el.Id == 0 {
	//  el.Id = util.AutoId.Get( consts.DB_OSM_FILE_POLYGONS_COLLECTIONS ) //fixme: o que esta constante está fazendo aqui nesse arquivo?
	//}

	var geoJSon GeoJSon = GeoJSon{}
	geoJSon.Init()
	geoJSon.AddGeoMathPolygonList(strconv.FormatInt(el.Id, 10), el)
	el.GeoJSonFeature, _ = geoJSon.StringLastFeature()

	return el.GeoJSonFeature
}

func (el *PolygonListStt) Initialize() {
	el.Initialized = true

	el.BBox = GetBoxPolygonList(el)
	el.BBoxBSon = GetBSonBoxInDegreesPolygonList(el)

}

func (el *PolygonListStt) ConvertToConvexHull() {
	var PointList PointListStt = PointListStt{}
	PointList.List = make([]PointStt, 0)

	for _, polygon := range el.List {
		for _, point := range polygon.PointsList {
			PointList.List = append(PointList.List, point)
		}
	}

	var hull PointListStt = PointList.ConvexHull()
	el.List = make([]PolygonStt, 1)
	el.List[0].PointsList = hull.List
}

func (el *PolygonListStt) ConvertToConcaveHull(n float64) {
	var PointList PointListStt = PointListStt{}
	PointList.List = make([]PointStt, 0)

	for _, polygon := range el.List {
		for _, point := range polygon.PointsList {
			PointList.List = append(PointList.List, point)
		}
	}

	var hull PointListStt = PointList.ConcaveHull(n)
	el.List = make([]PolygonStt, 1)
	el.List[0].PointsList = hull.List
}

func (el *PolygonListStt) ToJSon() ([]byte, error) {
	return bson.MarshalJSON(el)
}

func (el *PolygonListStt) ToReader() io.Reader {
	err, data := el.ToBSon()
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err)
	}

	return bytes.NewReader(data)
}

func (el *PolygonListStt) FromJSon(in []byte) error {
	return bson.UnmarshalJSON(in, el)
}

func (el *PolygonListStt) ToExternalFile(file *os.File, typeId []byte) error {
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

func (el *PolygonListStt) ToFile(file io.Writer) error {
	_, err := io.Copy(file, el.ToReader())
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err)
		return err
	}

	return nil
}

func (el *PolygonListStt) FromFile(file io.Reader) error {
	var bytesLBty []byte
	var bufferLObj *bytes.Buffer = bytes.NewBuffer(bytesLBty)

	_, err := io.Copy(bufferLObj, file)
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err)
		return err
	}

	err = el.FromBSon(bufferLObj.Bytes())
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err.Error())
		return err
	}

	return nil
}

func (el *PolygonListStt) ToFilePath(filePath string) error {
	var byteBSon []byte
	var err error

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

func (el *PolygonListStt) FromFilePath(filePath string) error {
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
