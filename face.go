package closest

import (
	"github.com/go-gl/mathgl/mgl64"
)

type face struct {
	indices [3]int
	measure Measure
}

func newFace(simplex []*vertex, indices [3]int) (theFace *face) {
	theFace = &face{
		indices: indices,
		measure: Measure{
			ConvexHulls: [2][]*mgl64.Vec3{
				{
					&mgl64.Vec3{},
				},
			},
		},
	}

	for _, index := range indices {
		newVertex := simplex[index].coordinate
		theFace.measure.ConvexHulls[1] = append(theFace.measure.ConvexHulls[1],
			&newVertex,
		)
	}

	theFace.measure.gjk()

	return
}

func (face face) getNormal(simplex []*vertex) mgl64.Vec3 {
	return simplex[face.indices[1]].coordinate.Sub(simplex[face.indices[0]].coordinate).Cross(
		simplex[face.indices[2]].coordinate.Sub(simplex[face.indices[0]].coordinate),
	)
}
