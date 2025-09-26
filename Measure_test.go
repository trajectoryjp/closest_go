package closest

import (
	"math/rand"
	"time"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/xieyuschen/deepcopy"

	"testing"
)

func TestMeasureNonnegativeDistance(t *testing.T) {
	testMeasureNonnegativeDistance(
		t,
		5.233333333333333,
		[]*mgl64.Vec3{
			{0.0, 5.5, 0.0},
			{2.3, 1.0, -2.0},
			{8.1, 4.0, 2.4},
			{4.3, 5.0, 2.2},
			{2.5, 1.0, 2.3},
			{7.1, 1.0, 2.4},
			{1.0, 1.5, 0.3},
			{3.3, 0.5, 0.3},
			{6.0, 1.4, 0.2},
		},
		[]*mgl64.Vec3{
			{0.0, -5.5, 0.0},
			{-4.0, 1.0, 5.0},
		},
	)
}

func TestMeasureNonnegativeDistance_Degeneration(t *testing.T) {
	testMeasureNonnegativeDistance(
		t,
		2.0,
		[]*mgl64.Vec3{
			{10, 10, 10},
			{93.76614808098593, 10, 10},
		},
		[]*mgl64.Vec3{
			{26.902334690093994, 7.686383247375488, 12},
			{30.745525360107422, 7.686383247375488, 12},
			{30.745525360107422, 11.529574871063232, 12},
			{26.902334690093994, 11.529574871063232, 12},
			{26.902334690093994, 7.686383247375488, 13},
			{30.745525360107422, 7.686383247375488, 13},
			{30.745525360107422, 11.529574871063232, 13},
			{26.902334690093994, 11.529574871063232, 13},
		},
	)
}

func TestMeasureNonnegativeDistance_OutOfTetrahedron(t *testing.T) {
	testMeasureNonnegativeDistance(
		t,
		478.65237808227545,
		[]*mgl64.Vec3{
			{24.80916023254391, -436.06686488070386, 1},
			{24.809160232543945, 149.8855333328247, 1},
		},
		[]*mgl64.Vec3{
			{503.46153831481934, 0, 0},
			{503.46153831481934, 299.7710666656494, 0},
			{503.46153831481934, 299.7710666656494, 2},
			{503.46153831481934, 0, 2},
		},
	)
}

func TestMeasureNonnegativeDistance_InOfTetrahedron(t *testing.T) {
	if testing.Short() {
		t.Skip("TODO: Make this test succeed.")
	}

	testMeasureNonnegativeDistance(
		t,
		0.0,
		[]*mgl64.Vec3{
			{9.809160232543945, 74.8855333328247, 1},
			{499.80916023254395, 74.8855333328247, 1},
		},
		[]*mgl64.Vec3{
			{103.76688194274902, 73.02115726470947, 1},
			{103.76688194274902, 73.02115726470947, 2},
			{103.76688194274902, 76.86437606811523, 2},
			{103.76688194274902, 76.86437606811523, 1},
		},
	)
}

func TestMeasureNonnegativeDistance_MinError(t *testing.T) {
	testMeasureNonnegativeDistance(
		t,
		53.29158003236736,
		[]*mgl64.Vec3{
			{
				231.13410161715001,
				42.359085964038968,
				8.2070553228259087,
			},
			{
				231.13428923673928,
				42.360740889096633,
				8.3670506989583373,
			},
		},
		[]*mgl64.Vec3{
			{
				1126.8901406135462,
				506.76397722481852,
				-991.48334605572745,
			},
			{
				-694.78953127471266,
				-318.69762289359494,
				-991.48334605572745,
			},
			{
				-694.78953127471266,
				-318.69762289359494,
				1008.5166539442725,
			},
			{
				1126.8901406135462,
				506.76397722481852,
				1008.5166539442725,
			},
		},
	)
}

func TestMeasureNonnegativeDistance_MinError2(t *testing.T) {
	testMeasureNonnegativeDistance(
		t,
		53.29158003236736,
		[]*mgl64.Vec3{
			{
				136.33086399999999,
				36.325947999999997,
				100,
			},
			{
				136.33014399999999,
				36.325048000000002,
				140,
			},
			{
				136.33086399999999,
				36.325947999999997,
				140,
			},
		},
		[]*mgl64.Vec3{
			{
				136.33020401000977,
				36.325602178745555,
				124,
			},
			{
				136.330246925354,
				36.325602178745555,
				124,
			},
			{
				136.330246925354,
				36.325567603588468,
				124,
			},
			{
				136.33020401000977,
				36.325567603588468,
				124,
			},
			{
				136.33020401000977,
				36.325602178745555,
				128,
			},
			{
				136.330246925354,
				36.325602178745555,
				128,
			},
			{
				136.330246925354,
				36.325567603588468,
				128,
			},
			{
				136.33020401000977,
				36.325567603588468,
				128,
			},
		},
	)
}

func testMeasureNonnegativeDistance(
	t *testing.T,
	correctDistance float64,
	convexHull0, convexHull1 []*mgl64.Vec3,
) {
	measure := Measure{
		ConvexHulls: [2][]*mgl64.Vec3{
			convexHull0,
			convexHull1,
		},
	}

	start := time.Now()
	measure.MeasureNonnegativeDistance()
	t.Log("Time: ", time.Since(start))
	if measure.Distance != correctDistance {
		t.Error("The distance: ", measure.Distance, " is different from the correct distance: ", correctDistance)
	}
}

func TestMeasureDistance(t *testing.T) {
	testMeasureDistance(
		t,
		-0.8135953914471573,
		[]*mgl64.Vec3{
			{0.0, 5.5, 0.0},
			{2.3, 1.0, -2.0},
			{8.1, 4.0, 2.4},
			{4.3, 5.0, 2.2},
			{2.5, 1.0, 2.3},
			{7.1, 1.0, 2.4},
			{1.0, 1.5, 0.3},
			{3.3, 0.5, 0.3},
			{6.0, 1.4, 0.2},
		},
		[]*mgl64.Vec3{
			{5.0, 6.0, -1.0},
			{-4.0, 1.0, 5.0},
		},
	)
}

func TestMeasureDistance_Geodetic(t *testing.T) {
	testMeasureDistance(
		t,
		-8.103902144849304e-05,
		[]*mgl64.Vec3{
			{136.243592, 36.294155, 0},
			{136.243591519521, 36.3058526069559, 0.132705141790211},
			{136.249286077761, 36.3058526238534, 0.153129168786108},
			{136.2492857044, 36.2941550169325, 0.0204240279272199},
			{136.243592, 36.294155, 99.9999999990687},
			{136.249285614983, 36.2941550169343, 100.020423707552},
			{136.249285988325, 36.3058524401501, 100.153126765043},
			{136.243591519529, 36.3058524232507, 100.13270305749},
		},
		[]*mgl64.Vec3{
			{136.24420166015625, 36.29409768373033, 12},
			{136.24420166015625, 36.29423604083452, 12},
			{136.2443733215332, 36.29423604083452, 12},
			{136.2443733215332, 36.29409768373033, 12},
			{136.24420166015625, 36.29409768373033, 28},
			{136.2443733215332, 36.29409768373033, 28},
			{136.2443733215332, 36.29423604083452, 28},
			{136.24420166015625, 36.29423604083452, 28},
		},
	)
}

func TestMeasureDistance_Geodetic2(t *testing.T) {
	if testing.Short() {
		t.Skip("TODO: Make this test succeed.")
	}

	testMeasureDistance(
		t,
		5.7316269682416994e-05,
		[]*mgl64.Vec3{
			{136.2436866760254, 36.293959326380744, 12},
			{136.24385833740234, 36.293959326380744, 12},
			{136.24385833740234, 36.29409768373033, 12},
			{136.2436866760254, 36.29409768373033, 12},
			{136.2436866760254, 36.293959326380744, 28},
			{136.24385833740234, 36.293959326380744, 28},
			{136.24385833740234, 36.29409768373033, 28},
			{136.2436866760254, 36.29409768373033, 28},
		},
		[]*mgl64.Vec3{
			{136.243592, 36.29415500000001, 0},
			{136.2493088026763, 36.29415500000001, 0},
			{136.2493088026763, 36.30588827924294, 0},
			{136.243592, 36.30588827924294, 0},
			{136.243592, 36.29415500000001, 100.15410614013672},
			{136.2493088026763, 36.29415500000001, 100.15410614013672},
			{136.2493088026763, 36.30588827924294, 100.15410614013672},
			{136.243592, 36.30588827924294, 100.15410614013672},
		},
	)
}

func TestMeasureDistance_DistanceNaN(t *testing.T) {
	convexHull0 := []*mgl64.Vec3{
		{0.8594475607808709, 0.9742341196245268, 0.03881845158332072},
		{0.11518805721821658, 0.2886100593167679, 0.7264075543605955},
	}

	convexHull1 := []*mgl64.Vec3{
		{0.1808976766933622, 0.4678535876991557, 0.39595195969136837},
		{0.9318649386849539, -0.061164616366541524, 0.12579316768712678},
		{0.3326005890627055, 0.053609576287277694, 0.7200526359540806},
		{0.147048080416384, 0.1043118025314802, 0.11557811629097817},
		{0.5917329252351495, 0.5148435176841939, 0.7696251459508143},
	}

	correctDistance := 0.0

	measure := Measure{
		ConvexHulls: [2][]*mgl64.Vec3{
			convexHull0,
			convexHull1,
		},
		Direction: mgl64.Vec3{-0.09838696251414104, 0.19117353980163715, 0.08413127327169329},
	}

	start := time.Now()
	measure.MeasureDistance()
	t.Log("Time: ", time.Since(start))

	if measure.Distance != correctDistance {
		t.Error("The distance: ", measure.Distance, " is different from the correct distance: ", correctDistance)
	}
}

func testMeasureDistance(
	t *testing.T,
	correctDistance float64,
	convexHull0, convexHull1 []*mgl64.Vec3,
) {
	measure := Measure{
		ConvexHulls: [2][]*mgl64.Vec3{
			convexHull0,
			convexHull1,
		},
	}

	start := time.Now()
	measure.MeasureDistance()
	t.Log("Time: ", time.Since(start))

	if measure.Distance != correctDistance {
		t.Error("The distance: ", measure.Distance, " is different from the correct distance: ", correctDistance)
	}
}

func TestMeasureDistanceRandomly(t *testing.T) {
	if testing.Short() {
		t.Skip("TODO: Make this test succeed.")
	}

	minDistance := 0.0
	tryCount := 0
	notCancelCount := 0
	var worstMeasure *Measure
	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		convexHull0 := make([]*mgl64.Vec3, rand.Intn(8)+1)
		for i := range convexHull0 {
			convexHull0[i] = &mgl64.Vec3{
				rand.Float64(),
				rand.Float64(),
				rand.Float64(),
			}
		}

		convexHull1 := make([]*mgl64.Vec3, rand.Intn(8)+1)
		for i := range convexHull1 {
			convexHull1[i] = &mgl64.Vec3{
				rand.Float64(),
				rand.Float64(),
				rand.Float64(),
			}
		}

		measure := Measure{
			ConvexHulls: [2][]*mgl64.Vec3{
				convexHull0,
				convexHull1,
			},
		}

		measure.MeasureDistance()

		if measure.Distance >= 0.0 {
			continue
		}

		it, _ := deepcopy.Copy(measure)
		shiftedMeasure := it.(Measure)

		for _, vertex := range shiftedMeasure.ConvexHulls[1] {
			*vertex = vertex.Sub(shiftedMeasure.Direction)
		}

		shiftedMeasure.MeasureDistance()
		tryCount += 1

		if shiftedMeasure.Distance < 0.0 {
			notCancelCount += 1
		}
		if shiftedMeasure.Distance >= minDistance {
			continue
		}
		worstMeasure = &measure
		minDistance = shiftedMeasure.Distance
	}

	t.Log("Try Count: ", tryCount)
	if worstMeasure != nil {
		t.Error("Not cancel collision!:",
			"\nNot Cancel Count: ", notCancelCount,
			"\nWorst Measure: ", worstMeasure,
		)
	}
}
