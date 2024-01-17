package closest

import (
	"github.com/go-gl/mathgl/mgl64"

	"math"
)

type vertex struct {
	indices               [2]int
	coordinate            mgl64.Vec3
	barycentricCoordinate float64
	isVisited             bool
}

func newVertex(convexes [2][]*mgl64.Vec3, direction mgl64.Vec3) *vertex {
	closestIndex0 := getIndexOfMaxDotWithDirection(convexes[0], direction)
	closestIndex1 := getIndexOfMaxDotWithDirection(convexes[1], direction.Mul(-1.0))

	return &vertex{
		indices: [2]int{
			closestIndex0,
			closestIndex1,
		},
		coordinate: convexes[1][closestIndex1].Sub(*convexes[0][closestIndex0]), // The dot product with direction is min
	}
}

func getIndexOfMaxDotWithDirection(convex []*mgl64.Vec3, direction mgl64.Vec3) (furthestIndex int) {
	maxS := math.Inf(-1.0)

	for i, vertex := range convex {
		s := vertex.Dot(direction)
		if s > maxS {
			furthestIndex = i
			maxS = s
		}
	}

	return
}
