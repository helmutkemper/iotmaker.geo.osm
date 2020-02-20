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
	"time"
)

type MembersStt struct {
	Type string `bson:"type"`
	Ref  int64  `bson:"ref"`
	Role string `bson:"role"`
}

type RelationStt struct {
	// id do open street maps
	Id             int64   `bson:"id"`
	IdWay          []int64 `bson:"idWay"`
	IdNode         []int64 `bson:"idNode"`
	IdMainRelation []int64 `bson:"idMainRelation"`
	IdRelation     []int64 `bson:"idRelation"`
	IdPolygon      []int64 `bson:"idPolygon"`
	IdPolygonsList []int64 `bson:"idPolygonsList"`
	Version        int64   `bson:"Version"`
	// TimeStamp dentro do Open Street Maps
	TimeStamp time.Time `bson:"timeStamp"`
	// ChangeSet dentro do Open Street Maps
	ChangeSet int64 `bson:"changeSet"`

	Visible bool

	// User Id dentro do Open Street Maps
	UId int64 `bson:"userId"`
	// User Name dentro do Open Street Maps
	User string `bson:"-"`
	// Tags do Open Street Maps
	// As Tags contêm _todo tipo de informação, desde como elas foram importadas, ao nome de um estabelecimento comercial,
	// por exemplo.
	Tag            map[string]string `bson:"tag"`
	International  map[string]string `bson:"international"`
	GeoJSon        string            `bson:"geoJSon"`
	GeoJSonFeature string            `bson:"geoJSonFeature"`

	Members []MembersStt `bson:"-"`

	Md5  [16]byte `bson:"md5" json:"-"`
	Size int      `bson:"size" json:"-"`

	HasKeyValue bool `bson:"hasKeyValue" json:"-"`
}

func (el *RelationStt) GetIdAsByte() []byte {
	var ret = make([]byte, 8)
	binary.LittleEndian.PutUint64(ret, uint64(el.Id))

	return ret
}

func (el *RelationStt) MakeMD5() (error, []byte) {
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

func (el *RelationStt) CheckMD5() error {
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

func (el *RelationStt) ToBSon() (error, []byte) {
	return el.MakeMD5()
}

func (el *RelationStt) FromBSon(byteBSon []byte) error {
	var err error

	err = bson.Unmarshal(byteBSon, el)
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err)
	}

	return err
}

func (el *RelationStt) ToJSon() ([]byte, error) {
	var err error
	var ret []byte

	ret, err = bson.MarshalJSON(el)
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypeRelation.error: %s", err.Error())
	}

	return ret, err
}

func (el *RelationStt) ToReader() io.Reader {
	err, data := el.ToBSon()
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypeRelation.error: %s", err.Error())
	}

	return bytes.NewReader(data)
}

func (el *RelationStt) FromJSon(in []byte) error {
	err := bson.UnmarshalJSON(in, el)
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypeRelation.error: %s", err.Error())
	}

	return err
}

func (el *RelationStt) ToFile(file io.Writer) error {
	_, err := io.Copy(file, el.ToReader())
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypeRelation.error: %s", err.Error())
		return err
	}

	return nil
}

func (el *RelationStt) FromFile(file io.Reader) error {
	var bytesLBty []byte
	var bufferLObj *bytes.Buffer = bytes.NewBuffer(bytesLBty)

	_, err := io.Copy(bufferLObj, file)
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypeRelation.error: %s", err.Error())
		return err
	}

	err = el.FromBSon(bufferLObj.Bytes())
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypeRelation.error: %s", err.Error())
	}

	return err
}

func (el *RelationStt) ToExternalFile(file *os.File, typeId []byte) error {
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

func (el *RelationStt) ToFilePath(filePath string) error {
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

func (el *RelationStt) FromFilePath(filePath string) error {
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
