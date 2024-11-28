// Package closest provides you calculating the closest points of two convex hulls.
// You get the distance or the depth between them in passing.
//
// The fundamental structure is [Measure]. An [Measure] contains two convex hulls, so
// you must set them at the first. Then, you can measure the distance or the depth between
// them by calling [Measure.MeasureDistance] or [Measure.MeasureNonnegativeDistance] with the closest points.
// You can reuse [Measure] any number of times. [Measure] stores the last
// direction from the first convex hull to the second convex hull, so it can calculate
// the closest points of the convex hulls faster than the first time.
package closest

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"

	"log"
	"sort"
)

// Measure is an all-in-one structure for calculating closest points of two convex hulls.
type Measure struct {
	// In
	// ConvexHulls are measured the distance between them.
	// The less degenerate the convex hull, the more precise the result.
	ConvexHulls [2][]*mgl64.Vec3

	// Out
	// Distance. If this is non-negative, this represents well-known distance s, (ds)² = (dx)² + (dy)² + (dz)².
	// This is more precise than Direction.Len().
	// If this is negative, this represents depth, which is the smallest distance for solving collisions.
	Distance float64
	// Direction is from ConvexHulls[0] to ConvexHulls[1].
	// This is more precise than Points[1].Sub(Points[0]).
	Direction mgl64.Vec3
	// Points are the closest points on each convex hulls.
	Points [2]mgl64.Vec3
	// Ons are the sets of indices of the vertices that make up the simplex that contains the closest point.
	Ons [2]map[int]struct{}

	simplex []*vertex
}

// MeasureDistance measures the distance or the depth between each ConvexHulls, and updates Direction, Points and Ons.
func (measure *Measure) MeasureDistance() {
	for _, convex := range measure.ConvexHulls {
		if len(convex) == 0 {
			measure.Distance = 0.0
			measure.Points = [2]mgl64.Vec3{}
			measure.Ons = [2]map[int]struct{}{
				{},
				{},
			}
			return
		}
	}
	measure.gjk()

	if len(measure.simplex) < 4 {
		return
	}

	measure.epa()
}

// MeasureNonnegativeDistance measures distance between each ConvexHulls, and updates Direction, Points and Ons.
func (measure *Measure) MeasureNonnegativeDistance() {
	for _, convex := range measure.ConvexHulls {
		if len(convex) == 0 {
			measure.Distance = 0.0
			measure.Points = [2]mgl64.Vec3{}
			measure.Ons = [2]map[int]struct{}{
				{},
				{},
			}
			return
		}
	}

	measure.gjk()
}

func (measure *Measure) gjk() {
	measure.simplex = measure.simplex[:0]
	measure.Distance = math.Inf(1)

	for len(measure.simplex) < 4 {
		measure.simplex = append(measure.simplex, newVertex(measure.ConvexHulls, measure.Direction))

		if measure.simplexHasCyclic(len(measure.simplex)-1, 0) {
			measure.simplex = measure.simplex[:len(measure.simplex)-1]
			break
		}

		if measure.updateSimplex() {
			measure.updateDirection()
			break
		}

		measure.updateDirection()
		lastDistance := measure.Distance
		measure.updateDistance()
		if measure.Distance >= lastDistance {
			break
		}
	}

	measure.updateTheOthers()
}

func (measure *Measure) epa() {
	// Distance　descending order
	faces := []*face{}
	switch len(measure.simplex) {
	case 3:
		newFace := newFace(measure.simplex, [3]int{0, 1, 2})
		if newFace.getNormal(measure.simplex).Dot(newFace.measure.Direction) < 0.0 {
			newFace.indices[1], newFace.indices[2] = newFace.indices[2], newFace.indices[1]
		}

		faces = append(faces, newFace)
	case 4:
		for i := 0; i < len(measure.simplex); i += 1 {
			indices := [3]int{}
			k := 0
			for j := 0; j < len(measure.simplex); j += 1 {
				if i == j {
					continue
				}
				indices[k] = j
				k += 1
			}
			newFace := newFace(measure.simplex, indices)
			if newFace.getNormal(measure.simplex).Dot(newFace.measure.Direction) < 0.0 {
				newFace.indices[1], newFace.indices[2] = newFace.indices[2], newFace.indices[1]
			}

			faces = append(faces, newFace)
		}
	default:
		log.Panic("Must not come here!")
	}

	sort.Slice(faces, func(i int, j int) bool {
		return faces[i].measure.Distance > faces[j].measure.Distance
	})

findOuterMinDistanceFace:
	for {
		newVertex := newVertex(measure.ConvexHulls, faces[len(faces)-1].measure.Direction.Mul(-1))
		for _, vertex := range measure.simplex {
			if newVertex.indices == vertex.indices {
				break findOuterMinDistanceFace
			}
		}

		measure.simplex = append(measure.simplex, newVertex)
		faces = measure.reconstruct(faces)
	}

	newSimplex := []*vertex{}
	for _, index := range faces[len(faces)-1].indices {
		newSimplex = append(newSimplex, measure.simplex[index])
	}
	measure.simplex = newSimplex
	measure.updateSimplex()
	measure.updateDirection()
	measure.updateDistance()
	measure.updateTheOthers()

	measure.Distance *= -1.0
}

func (measure *Measure) simplexHasCyclic(i int, j int) bool {
	for newI := 0; newI < len(measure.simplex)-1; newI += 1 {
		if measure.simplex[newI].isVisited {
			continue
		}
		if measure.simplex[newI].indices[j] != measure.simplex[i].indices[j] {
			continue
		}

		measure.simplex[newI].isVisited = true
		newJ := (j + 1) % 2

		if measure.simplex[len(measure.simplex)-1].indices[newJ] == measure.simplex[newI].indices[newJ] {
			measure.simplex[newI].isVisited = false
			return true
		}
		if measure.simplexHasCyclic(newI, newJ) {
			measure.simplex[newI].isVisited = false
			return true
		}

		measure.simplex[newI].isVisited = false
	}

	return false
}

func (measure *Measure) updateSimplex() (isDegenerated bool) {
	switch len(measure.simplex) {
	case 1:
		measure.simplex[0].barycentricCoordinate = 1.0
	case 2:
		a := measure.simplex[0].coordinate
		b := measure.simplex[1].coordinate

		ab := b.Sub(a)

		u := b.Dot(ab)
		if u <= 0.0 {
			// Region B
			measure.simplex[0] = measure.simplex[1]
			measure.simplex = measure.simplex[:1]
			measure.simplex[0].barycentricCoordinate = 1.0
			break
		}

		// Region AB
		v := -a.Dot(ab)

		measure.simplex[0].barycentricCoordinate = u
		measure.simplex[1].barycentricCoordinate = v
	case 3:
		a := measure.simplex[0].coordinate
		b := measure.simplex[1].coordinate
		c := measure.simplex[2].coordinate

		ab := b.Sub(a)
		ac := c.Sub(a)
		bc := c.Sub(b)

		uBC := c.Dot(bc)
		uAC := c.Dot(ac)

		if uBC <= 0.0 && uAC <= 0.0 {
			// Region C
			measure.simplex[0] = measure.simplex[2]
			measure.simplex = measure.simplex[:1]

			measure.simplex[0].barycentricCoordinate = 1.0
			break
		}

		vBC := -b.Dot(bc)

		n := ab.Cross(ac)
		if n[0] == 0.0 && n[1] == 0.0 && n[2] == 0.0 {
			isDegenerated = true
			return
		}
		n1 := b.Cross(c)

		uABC := n1.Dot(n)

		if uABC <= 0.0 && uBC > 0.0 && vBC > 0.0 {
			// Region BC
			measure.simplex[0] = measure.simplex[1]
			measure.simplex[1] = measure.simplex[2]
			measure.simplex = measure.simplex[:2]

			measure.simplex[0].barycentricCoordinate = uBC
			measure.simplex[1].barycentricCoordinate = vBC
			break
		}

		vAC := -a.Dot(ac)

		n2 := c.Cross(a)

		vABC := n2.Dot(n)

		if vABC <= 0.0 && uAC > 0.0 && vAC > 0.0 {
			// Region AC
			// measure.simplex[0] = measure.simplex[0]
			measure.simplex[1] = measure.simplex[2]
			measure.simplex = measure.simplex[:2]

			measure.simplex[0].barycentricCoordinate = uAC
			measure.simplex[1].barycentricCoordinate = vAC
			break
		}

		// Region ABC
		n3 := a.Cross(b)

		wABC := n3.Dot(n)

		measure.simplex[0].barycentricCoordinate = uABC
		measure.simplex[1].barycentricCoordinate = vABC
		measure.simplex[2].barycentricCoordinate = wABC
	case 4:
		a := measure.simplex[0].coordinate
		b := measure.simplex[1].coordinate
		c := measure.simplex[2].coordinate
		d := measure.simplex[3].coordinate

		ad := d.Sub(a)
		bd := d.Sub(b)
		cd := d.Sub(c)

		uBD := d.Dot(bd)
		uCD := d.Dot(cd)
		uAD := d.Dot(ad)

		if uBD <= 0.0 && uCD <= 0.0 && uAD <= 0.0 {
			// Region D
			measure.simplex[0] = measure.simplex[3]
			measure.simplex = measure.simplex[:1]

			measure.simplex[0].barycentricCoordinate = 1.0
			break
		}

		ab := b.Sub(a)
		ac := c.Sub(a)
		bc := c.Sub(b)

		vBD := -b.Dot(bd)
		vCD := -c.Dot(cd)
		vAD := -a.Dot(ad)

		n := ad.Cross(ab)
		n1 := d.Cross(b)
		n2 := b.Cross(a)
		n3 := a.Cross(d)

		uADB := n1.Dot(n)
		vADB := n2.Dot(n)
		wADB := n3.Dot(n)

		n = ac.Cross(ad)
		n1 = c.Cross(d)
		n2 = d.Cross(a)
		n3 = a.Cross(c)

		uACD := n1.Dot(n)
		vACD := n2.Dot(n)
		wACD := n3.Dot(n)

		n = bc.Mul(-1.0).Cross(cd)
		n1 = b.Cross(d)
		n2 = d.Cross(c)
		n3 = c.Cross(b)

		uCBD := n1.Dot(n)
		vCBD := n2.Dot(n)
		wCBD := n3.Dot(n)

		if vCBD <= 0.0 && uACD <= 0.0 && uCD > 0.0 && vCD > 0.0 {
			// region DC
			measure.simplex[0] = measure.simplex[2]
			measure.simplex[1] = measure.simplex[3]
			measure.simplex = measure.simplex[:2]

			measure.simplex[0].barycentricCoordinate = uCD
			measure.simplex[1].barycentricCoordinate = vCD
			break
		}

		if vACD <= 0.0 && wADB <= 0.0 && uAD > 0.0 && vAD > 0.0 {
			// region AD
			// measure.simplex[0] = measure.simplex[0]
			measure.simplex[1] = measure.simplex[3]
			measure.simplex = measure.simplex[:2]

			measure.simplex[0].barycentricCoordinate = uAD
			measure.simplex[1].barycentricCoordinate = vAD
			break
		}

		if uCBD <= 0.0 && uADB <= 0.0 && uBD > 0.0 && vBD > 0.0 {
			//region BD
			measure.simplex[0] = measure.simplex[1]
			measure.simplex[1] = measure.simplex[3]
			measure.simplex = measure.simplex[:2]

			measure.simplex[0].barycentricCoordinate = uBD
			measure.simplex[1].barycentricCoordinate = vBD
			break
		}

		volume := -bc.Cross(ab).Dot(bd)
		if volume == 0.0 {
			isDegenerated = true
			return
		}
		volumeInverse := 1.0 / volume
		uABCD := c.Cross(d).Dot(b) * volumeInverse
		if uABCD == 0.0 {
			isDegenerated = true
			return
		}
		vABCD := c.Cross(a).Dot(d) * volumeInverse
		if vABCD == 0.0 {
			isDegenerated = true
			return
		}
		wABCD := d.Cross(a).Dot(b) * volumeInverse
		if wABCD == 0.0 {
			isDegenerated = true
			return
		}
		xABCD := b.Cross(a).Dot(c) * volumeInverse
		if xABCD == 0.0 {
			isDegenerated = true
			return
		}

		if uABCD < 0.0 && uCBD > 0.0 && vCBD > 0.0 && wCBD > 0.0 {
			// region CBD
			measure.simplex[0] = measure.simplex[1]
			measure.simplex[1] = measure.simplex[2]
			measure.simplex[2] = measure.simplex[3]
			measure.simplex = measure.simplex[:3]

			measure.simplex[0].barycentricCoordinate = vCBD
			measure.simplex[1].barycentricCoordinate = uCBD
			measure.simplex[2].barycentricCoordinate = wCBD
			break
		}
		if vABCD < 0.0 && uACD > 0.0 && vACD > 0.0 && wACD > 0.0 {
			// region ACD
			// measure.simplex[0] = measure.simplex[0]
			measure.simplex[1] = measure.simplex[2]
			measure.simplex[2] = measure.simplex[3]
			measure.simplex = measure.simplex[:3]

			measure.simplex[0].barycentricCoordinate = uACD
			measure.simplex[1].barycentricCoordinate = vACD
			measure.simplex[2].barycentricCoordinate = wACD
			break
		}
		if wABCD < 0.0 && uADB > 0.0 && vADB > 0.0 && wADB > 0.0 {
			// region ADB
			// measure.simplex[0] = measure.simplex[0]
			// measure.simplex[1] = measure.simplex[1]
			measure.simplex[2] = measure.simplex[3]
			measure.simplex = measure.simplex[:3]

			measure.simplex[0].barycentricCoordinate = uADB
			measure.simplex[1].barycentricCoordinate = wADB
			measure.simplex[2].barycentricCoordinate = vADB
			break
		}

		// region ABCD
		measure.simplex[0].barycentricCoordinate = uABCD
		measure.simplex[1].barycentricCoordinate = vABCD
		measure.simplex[2].barycentricCoordinate = wABCD
		measure.simplex[3].barycentricCoordinate = xABCD
	default:
		log.Panic("Must not come here!")
	}

	isDegenerated = false
	return
}

func (measure *Measure) updateDirection() {
	switch len(measure.simplex) {
	case 1:
		measure.Direction = measure.simplex[0].coordinate
	case 2:
		difference := measure.simplex[0].coordinate.Sub(measure.simplex[1].coordinate)

		x := measure.simplex[0].coordinate.Dot(difference) / difference.LenSqr()
		difference = difference.Mul(x)
		measure.Direction = measure.simplex[0].coordinate.Sub(difference)
	case 3:
		ba := measure.simplex[0].coordinate.Sub(measure.simplex[1].coordinate)
		ca := measure.simplex[0].coordinate.Sub(measure.simplex[2].coordinate)
		n := ba.Cross(ca)
		scale := n.Dot(measure.simplex[0].coordinate) / n.LenSqr()

		measure.Direction = n.Mul(scale)
	case 4:
		measure.Direction = mgl64.Vec3{}
	}
}

func (measure *Measure) updateDistance() {
	measure.Distance = measure.Direction.Len()
}

func (measure *Measure) updateTheOthers() {
	denominator := 0.0
	for _, vertex := range measure.simplex {
		denominator += vertex.barycentricCoordinate
	}
	denominator = 1.0 / denominator

	measure.Points = [2]mgl64.Vec3{}
	for i := 0; i < len(measure.Points); i += 1 {
		for _, vertex := range measure.simplex {
			measure.Points[i] = measure.Points[i].Add(measure.ConvexHulls[i][vertex.indices[i]].Mul(denominator * vertex.barycentricCoordinate))
		}
	}

	measure.Ons = [2]map[int]struct{}{
		{},
		{},
	}
	for i := 0; i < len(measure.Points); i += 1 {
		for _, vertex := range measure.simplex {
			measure.Ons[i][vertex.indices[i]] = struct{}{}
		}
	}
}

func (measure *Measure) reconstruct(faces []*face) []*face {
	reconstructEdges := map[[2]int]struct{}{}

	for i := len(faces) - 1; i >= 0; i-- {
		if faces[i].getNormal(measure.simplex).Dot(
			measure.simplex[len(measure.simplex)-1].coordinate.Sub(measure.simplex[faces[i].indices[0]].coordinate),
		) <= 0.0 { // If new simplex is below the face
			continue
		}

		for j := 0; j < 3; j += 1 {
			k := (j + 1) % 3
			edgeIndices := [2]int{faces[i].indices[j], faces[i].indices[k]}
			_, ok := reconstructEdges[edgeIndices]

			if ok {
				delete(reconstructEdges, edgeIndices)
			} else {
				reconstructEdges[[2]int{edgeIndices[1], edgeIndices[0]}] = struct{}{}
			}
		}

		faces = append(faces[:i], faces[i+1:]...)
	}

reconstruct:
	for edge := range reconstructEdges {
		newFace := newFace(measure.simplex, [3]int{
			edge[1],
			edge[0],
			len(measure.simplex) - 1,
		})

		// Append the new face to the faces
		for i := len(faces) - 1; i >= 0; i-- {
			if newFace.measure.Distance <= faces[i].measure.Distance {
				faces = append(faces[:i+1], faces[i:]...)
				faces[i+1] = newFace
				continue reconstruct
			}
		}

		faces = append(faces[:1], faces[:]...)
		faces[0] = newFace
	}

	return faces
}
