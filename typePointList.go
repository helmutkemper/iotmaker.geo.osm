package iotmaker_geo_osm

import (
	"crypto/md5"
	"encoding/binary"
	"errors"
	"github.com/helmutkemper/mgo/bson"
	log "github.com/helmutkemper/seelog"
)

// Point list for find multiples points into db.
type PointListStt struct {
	// id do open street maps
	Id   int64      `bson:"id"`
	List []PointStt `bson:"list"`

	Md5  [16]byte `bson:"md5" json:"-"`
	Size int      `bson:"size" json:"-"`

	HasKeyValue bool `bson:"hasKeyValue" json:"-"`
}

func (el *PointListStt) GetId() int64 {
	return el.Id
}

func (el *PointListStt) GetIdAsByte() []byte {
	var ret = make([]byte, 8)
	binary.LittleEndian.PutUint64(ret, uint64(el.Id))

	return ret
}

func (el *PointListStt) MakeMD5() (error, []byte) {
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

func (el *PointListStt) CheckMD5() error {
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

func (el *PointListStt) ToBSon() (error, []byte) {
	return el.MakeMD5()
}

func (el *PointListStt) FromBSon(byteBSon []byte) error {
	var err error

	err = bson.Unmarshal(byteBSon, el)
	if err != nil {
		log.Criticalf("gOsm.geoMath.geoTypePolygon.error: %s", err)
	}

	return err
}

func (el *PointListStt) AddFromPointList(point *PointListStt) {
	if len(el.List) == 0 {
		el.List = make([]PointStt, 0)
	}

	for _, p := range point.List {
		el.List = append(el.List, p)
	}
}

func (el *PointListStt) AddPointLatLngDegrees(latitudeAFlt, longitudeAFlt float64) error {
	if len(el.List) == 0 {
		el.List = make([]PointStt, 0)
	}

	var p = PointStt{}
	var err = p.SetLatLngDegrees(latitudeAFlt, longitudeAFlt)
	if err != nil {
		return err
	}

	el.List = append(el.List, p)

	return nil
}

func (el *PointListStt) AddPointLatLngDecimalDrees(latitudeDegreesAFlt, latitudePrimesAFlt, latitudeSecondsAFlt, longitudeDegreesAFlt, longitudePrimesAFlt, longitudeSecondsAFlt int64) {
	if len(el.List) == 0 {
		el.List = make([]PointStt, 0)
	}

	var p = PointStt{}
	p.SetLatLngDecimalDrees(latitudeDegreesAFlt, latitudePrimesAFlt, latitudeSecondsAFlt, longitudeDegreesAFlt, longitudePrimesAFlt, longitudeSecondsAFlt)
	el.List = append(el.List, p)
}

func (el *PointListStt) AddPointLngLatDecimalDrees(longitudeDegreesAFlt, longitudePrimesAFlt, longitudeSecondsAFlt, latitudeDegreesAFlt, latitudePrimesAFlt, latitudeSecondsAFlt int64) {
	if len(el.List) == 0 {
		el.List = make([]PointStt, 0)
	}

	var p = PointStt{}
	p.SetLngLatDecimalDrees(longitudeDegreesAFlt, longitudePrimesAFlt, longitudeSecondsAFlt, latitudeDegreesAFlt, latitudePrimesAFlt, latitudeSecondsAFlt)
	el.List = append(el.List, p)
}

func (el *PointListStt) AddPointLngLatDegrees(longitudeAFlt, latitudeAFlt float64) error {
	if len(el.List) == 0 {
		el.List = make([]PointStt, 0)
	}

	var p = PointStt{}
	var err = p.SetLngLatDegrees(longitudeAFlt, latitudeAFlt)
	if err != nil {
		return err
	}

	el.List = append(el.List, p)

	return nil
}

func (el *PointListStt) AddPointXYDegrees(xAFlt, yAFlt float64) error {
	if len(el.List) == 0 {
		el.List = make([]PointStt, 0)
	}

	var p = PointStt{}
	var err = p.SetXYDegrees(xAFlt, yAFlt)
	if err != nil {
		return err
	}

	el.List = append(el.List, p)

	return nil
}

func (el *PointListStt) AddPointLatLngRadians(latitudeAFlt, longitudeAFlt float64) error {
	if len(el.List) == 0 {
		el.List = make([]PointStt, 0)
	}

	var p = PointStt{}
	var err = p.SetLatLngRadians(latitudeAFlt, longitudeAFlt)
	if err != nil {
		return err
	}

	el.List = append(el.List, p)

	return nil
}

func (el *PointListStt) AddPointLngLatRadians(longitudeAFlt, latitudeAFlt float64) error {
	if len(el.List) == 0 {
		el.List = make([]PointStt, 0)
	}

	var p = PointStt{}
	var err = p.SetLngLatRadians(longitudeAFlt, latitudeAFlt)
	if err != nil {
		return err
	}

	el.List = append(el.List, p)

	return nil
}

func (el *PointListStt) AddPointXYRadians(xAFlt, yAFlt float64) error {
	if len(el.List) == 0 {
		el.List = make([]PointStt, 0)
	}

	var p = PointStt{}
	var err = p.SetXYRadians(xAFlt, yAFlt)
	if err != nil {
		return err
	}

	el.List = append(el.List, p)

	return nil
}

// Get all points after Find()
func (el *PointListStt) GetAll() []PointStt {
	return el.List
}

func (el *PointListStt) GetBox() BoxStt {
	return GetBox(&el.List)
}

func (el *PointListStt) GetBSonBoxInDegrees() bson.M {
	return GetBSonBoxInDegrees(&el.List)
}

func (el PointListStt) GetReverse() PointListStt {
	for left, right := 0, len(el.List)-1; left < right; left, right = left+1, right-1 {
		el.List[left], el.List[right] = el.List[right], el.List[left]
	}

	return el
}

func (el PointListStt) AddPoint(PointAStt PointStt) {
	el.List = append(el.List, PointAStt)
}

func (el PointListStt) hullIsPointInsidePolygon(pointBStt PointStt, polygonAAStt []PointStt) bool {
	var i int
	var result = false
	var j = len(polygonAAStt) - 1

	for i = 0; i < len(polygonAAStt); i += 1 {
		if (polygonAAStt[i].Loc[1] < pointBStt.Loc[1] && polygonAAStt[j].Loc[1] > pointBStt.Loc[1]) || (polygonAAStt[j].Loc[1] < pointBStt.Loc[1] && polygonAAStt[i].Loc[1] > pointBStt.Loc[1]) {
			if polygonAAStt[i].Loc[0]+(pointBStt.Loc[1]-polygonAAStt[i].Loc[1])/(polygonAAStt[j].Loc[1]-polygonAAStt[i].Loc[1])*(polygonAAStt[j].Loc[0]-polygonAAStt[i].Loc[0]) < pointBStt.Loc[0] {
				result = !result
			}
		}
		j = i
	}

	return result
}

func (el PointListStt) hullCheckEdgeIntersection(p0, p1, p2, p3 PointStt) bool {
	var s1_x = p1.Loc[0] - p0.Loc[0]
	var s1_y = p1.Loc[1] - p0.Loc[1]
	var s2_x = p3.Loc[0] - p2.Loc[0]
	var s2_y = p3.Loc[1] - p2.Loc[1]
	var s = (-s1_y*(p0.Loc[0]-p2.Loc[0]) + s1_x*(p0.Loc[1]-p2.Loc[1])) / (-s2_x*s1_y + s1_x*s2_y)
	var t = (s2_x*(p0.Loc[1]-p2.Loc[1]) - s2_y*(p0.Loc[0]-p2.Loc[0])) / (-s2_x*s1_y + s1_x*s2_y)

	return s > 0 && s < 1 && t > 0 && t < 1
}

func (el PointListStt) hullCheckEdgeIntersectionList(hull []PointStt, curEdgeStart, curEdgeEnd, checkEdgeStart, checkEdgeEnd PointStt) bool {
	var i int
	var e1, e2 int
	var p1, p2 PointStt
	for i = 0; i < len(hull)-2; i += 1 {
		e1 = i
		e2 = i + 1
		p1 = hull[e1]
		p2 = hull[e2]

		if curEdgeStart.Equality(p1) && curEdgeEnd.Equality(p2) {
			continue
		}

		if el.hullCheckEdgeIntersection(checkEdgeStart, checkEdgeEnd, p1, p2) {
			return true
		}
	}
	return false
}

func (el *PointListStt) ConvertToConvexHull() {
	var hull = el.ConvexHull()
	el.List = hull.List
}

func (el PointListStt) ConvexHull() PointListStt {
	var i, j, bot int
	var tmp = PointStt{}
	var P = make([]PointStt, len(el.List)) // = el.List
	var hull = PointListStt{}
	hull.List = make([]PointStt, 0)
	var minmin, minmax int
	var maxmin, maxmax int
	var xmin, xmax float64

	for k, v := range el.List {
		P[k].CopyFrom(v)
	}

	// Sort P by x and y
	for i = 0; i < len(P); i += 1 {
		for j = i + 1; j < len(P); j += 1 {
			if P[j].Loc[0] < P[i].Loc[0] || (P[j].Loc[0] == P[i].Loc[0] && P[j].Loc[1] < P[i].Loc[1]) {
				tmp.CopyFrom(P[i])
				P[i].CopyFrom(P[j])
				P[j].CopyFrom(tmp)
			}
		}
	}

	// the output array H[] will be used as the stack
	// i array scan index

	// Get the indices of points with min x-coord and min|max y-coord
	minmin = 0
	xmin = P[0].Loc[0]
	for i = 1; i < len(P); i += 1 {
		if P[i].Loc[0] != xmin {
			break
		}
	}

	minmax = i - 1
	if minmax == len(P)-1 { // degenerate case: all x-coords == xmin
		hull.List = append(hull.List, P[minmin])
		if P[minmax].Loc[1] != P[minmin].Loc[1] {
			hull.List = append(hull.List, P[minmax]) // a  nontrivial segment
		}
		hull.List = append(hull.List, P[minmin]) // add polygon endpoint
		return hull
	}

	// Get the indices of points with max x-coord and min|max y-coord
	maxmax = len(P) - 1
	xmax = P[len(P)-1].Loc[0]
	for i = len(P) - 2; i >= 0; i -= 1 {
		if P[i].Loc[0] != xmax {
			break
		}
	}
	maxmin = i + 1

	// Compute the lower hull on the stack H
	hull.List = append(hull.List, P[minmin]) // push  minmin point onto stack
	i = minmax
	for i+1 <= maxmin {
		i += 1

		// the lower line joins P[minmin]  with P[maxmin]
		if el.hullCcw(P[minmin], P[maxmin], P[i]) >= 0 && i < maxmin {
			continue // ignore P[i] above or on the lower line
		}

		for len(hull.List) > 1 { // there are at least 2 points on the stack
			// test if  P[i] is left of the line at the stack top
			if el.hullCcw(hull.List[len(hull.List)-2], hull.List[len(hull.List)-1], P[i]) > 0 {
				break // P[i] is a new hull  vertex
			}
			hull.List = hull.List[:len(hull.List)-1] // pop top point off  stack
		}
		hull.List = append(hull.List, P[i]) // push P[i] onto stack
	}

	// Next, compute the upper hull on the stack H above  the bottom hull
	if maxmax != maxmin { // if  distinct xmax points
		hull.List = append(hull.List, P[maxmax]) // push maxmax point onto stack
	}
	bot = len(hull.List) // the bottom point of the upper hull stack
	i = maxmin
	for (i - 1) >= minmax {
		i -= 1
		// the upper line joins P[maxmax]  with P[minmax]
		if el.hullCcw(P[maxmax], P[minmax], P[i]) >= 0 && i > minmax {
			continue // ignore P[i] below or on the upper line
		}

		for len(hull.List) > bot { // at least 2 points on the upper stack
			// test if  P[i] is left of the line at the stack top
			if el.hullCcw(hull.List[len(hull.List)-2], hull.List[len(hull.List)-1], P[i]) > 0 {
				break // P[i] is a new hull  vertex
			}

			hull.List = hull.List[:len(hull.List)-1] // pop top point off stack
		}
		hull.List = append(hull.List, P[i]) // push P[i] onto stack
	}
	if minmax != minmin {
		hull.List = append(hull.List, P[minmin]) // push  joining endpoint onto stack
	}

	return hull
}

func (el *PointListStt) ConvertToConcaveHull(n float64) {
	var hull = el.ConcaveHull(n)
	el.List = hull.List
}

func (el PointListStt) ConcaveHull(n float64) PointListStt {
	var i, z int
	var eh, dd float64
	var found, intersects bool
	var ci1, ci2, pk PointStt
	var tmp = make([]PointStt, 0)
	var hull = el.ConvexHull()

	var d, dTmp float64
	var skip bool
	var distance = 0.0

	for i = 0; i < len(hull.List)-1; i += 1 {
		// Find the nearest inner point pk ∈ G from the edge (ci1, ci2);
		ci1.CopyFrom(hull.List[i])
		ci2.CopyFrom(hull.List[i+1])

		distance = 0.0
		found = false

		for _, p := range el.List {
			// Skip points that are already in he hull
			if p.IsContainedInTheList(hull.List) {
				continue
			}

			d = p.Distance(ci1, ci2)
			skip = false
			for z = 0; !skip && z < len(hull.List)-1; z += 1 {
				dTmp = p.Distance(hull.List[z], hull.List[z+1])
				skip = skip || dTmp < d
			}
			if skip {
				continue
			}

			if !found || distance > d {
				pk = p
				distance = d
				found = true
			}
		}

		if !found || pk.IsContainedInTheList(hull.List) {
			continue
		}

		eh = ci1.Pythagoras(ci2) // the lenght of the edge
		tmp = make([]PointStt, 0)
		tmp = append(tmp, ci1)
		tmp = append(tmp, ci2)

		dd = pk.DecisionDistance(tmp)

		if eh/dd > n {
			// Check that new candidate edge will not intersect existing edges.
			intersects = el.hullCheckEdgeIntersectionList(hull.List, ci1, ci2, ci1, pk)
			intersects = intersects || el.hullCheckEdgeIntersectionList(hull.List, ci1, ci2, pk, ci2)
			if !intersects {
				hull.List = append(hull.List[:(i+1)], append([]PointStt{pk}, hull.List[(i+1):]...)...)
				i -= 1
			}
		}
	}

	return hull
}

func (el PointListStt) hullCcw(p1, p2, p3 PointStt) float64 {
	return (p2.Rad[0]-p1.Rad[0])*(p3.Rad[1]-p1.Rad[1]) - (p2.Rad[1]-p1.Rad[1])*(p3.Rad[0]-p1.Rad[0])
}
