package sim

import (
	"math"
	"math/rand"

	"github.com/aliaksei135/abs-specific/hist"

	"gonum.org/v1/gonum/mat"
)

const (
	RAD2DEG = 180.0 / math.Pi
	DEG2RAD = math.Pi / 180
)

func bearing2angle(bearing float64) float64 {
	return math.Mod((360 - (bearing - 90)), 360)
}

type Traffic struct {
	//Setup
	x_bounds      [2]float64
	y_bounds      [2]float64
	z_bounds      [2]float64
	target_agents int
	TotalVolume   float64

	//Randomness
	AltitudeDistr     hist.Histogram
	VelocityDistr     hist.Histogram
	TrackDistr        hist.Histogram
	VerticalRateDistr hist.Histogram
	SurfaceEntrance   bool
	randomSource      *rand.Rand

	//State
	StepVelocities mat.Dense
	Positions      mat.Dense
	Seed           int64
	Timestep       float64
	oob_rows       []int
}

func (tfc *Traffic) randFloat64() float64 {
	return float64(tfc.randomSource.Int63n(1<<53)) / (1 << 53)
}

func (tfc *Traffic) Setup(bounds [6]float64, target_density float64) {

	tfc.x_bounds[0] = bounds[0] - 1000
	tfc.x_bounds[1] = bounds[1] + 1000
	tfc.y_bounds[0] = bounds[2] - 1000
	tfc.y_bounds[1] = bounds[3] + 1000
	tfc.z_bounds[0] = bounds[4] - 200
	tfc.z_bounds[1] = bounds[5] + 200

	tfc.randomSource = rand.New(rand.NewSource(tfc.Seed))

	tfc.TotalVolume = math.Abs(tfc.x_bounds[1]-tfc.x_bounds[0]) * math.Abs(tfc.y_bounds[1]-tfc.y_bounds[0]) * math.Abs(tfc.z_bounds[1]-tfc.z_bounds[0])
	tfc.target_agents = int(math.Ceil(target_density * tfc.TotalVolume))

	tfc.oob_rows = make([]int, tfc.target_agents)
	tfc.Positions = *mat.NewDense(tfc.target_agents, 3, nil)
	tfc.StepVelocities = *mat.NewDense(tfc.target_agents, 3, nil)

	for i := range tfc.oob_rows {
		tfc.oob_rows[i] = i
	}
	tfc.AddAgents()
}

func (tfc *Traffic) GenerateXYEdgePosition() [2]float64 {
	x_pos := ((tfc.x_bounds[1] - tfc.x_bounds[0]) * tfc.randFloat64()) + tfc.x_bounds[0]
	y_pos := ((tfc.y_bounds[1] - tfc.y_bounds[0]) * tfc.randFloat64()) + tfc.y_bounds[0]

	if tfc.SurfaceEntrance {
		switch r := tfc.randFloat64(); {
		case r < 0.25:
			x_pos = tfc.x_bounds[0]
		case r < 0.5:
			x_pos = tfc.x_bounds[1]
		case r < 0.75:
			y_pos = tfc.y_bounds[0]
		default:
			y_pos = tfc.y_bounds[1]
		}
	}

	return [2]float64{x_pos, y_pos}
}

func (tfc *Traffic) AddAgents() {
	n_new_agents := len(tfc.oob_rows)
	speeds := tfc.VelocityDistr.Sample(n_new_agents, *tfc.randomSource)
	tracks := tfc.TrackDistr.Sample(n_new_agents, *tfc.randomSource)
	vert_rates := tfc.VerticalRateDistr.Sample(n_new_agents, *tfc.randomSource)
	alts := tfc.AltitudeDistr.Sample(n_new_agents, *tfc.randomSource)
	for idx, insert_row_idx := range tfc.oob_rows {
		xy_pos := tfc.GenerateXYEdgePosition()
		z_pos := alts[idx]
		tfc.Positions.Set(insert_row_idx, 0, xy_pos[0])
		tfc.Positions.Set(insert_row_idx, 1, xy_pos[1])
		tfc.Positions.Set(insert_row_idx, 2, z_pos)

		x_vel := math.Cos(DEG2RAD*bearing2angle(tracks[idx])) * speeds[idx]
		y_vel := math.Sin(DEG2RAD*bearing2angle(tracks[idx])) * speeds[idx]
		z_vel := vert_rates[idx]
		tfc.StepVelocities.Set(insert_row_idx, 0, x_vel*tfc.Timestep)
		tfc.StepVelocities.Set(insert_row_idx, 1, y_vel*tfc.Timestep)
		tfc.StepVelocities.Set(insert_row_idx, 2, z_vel*tfc.Timestep)
	}

	tfc.oob_rows = tfc.oob_rows[:0] // Clear filled oob rows
}

func (tfc *Traffic) Step() {
	tfc.Positions.Add(&tfc.Positions, &tfc.StepVelocities)

	for i := 0; i < tfc.Positions.RawMatrix().Rows; i++ {
		if tfc.Positions.At(i, 0) < tfc.x_bounds[0] || tfc.Positions.At(i, 0) > tfc.x_bounds[1] || tfc.Positions.At(i, 1) < tfc.y_bounds[0] || tfc.Positions.At(i, 1) > tfc.y_bounds[1] || tfc.Positions.At(i, 2) < tfc.z_bounds[0] || tfc.Positions.At(i, 2) > tfc.z_bounds[1] {
			tfc.oob_rows = append(tfc.oob_rows, i)
		}
	}

	if len(tfc.oob_rows) > 1 {
		tfc.AddAgents()
	}
}

func (tfc *Traffic) End() {

}

type Ownship struct {
	Path         [][3]float64
	position     [3]float64
	Velocity     float64
	Timestep     float64
	pathIndex    int
	stepVelocity float64
}

func (ownship *Ownship) Setup() {
	ownship.pathIndex = 1
	ownship.position = ownship.Path[0]
	ownship.stepVelocity = ownship.Velocity * ownship.Timestep
}

func (ownship *Ownship) Step() {
	sub_goal := ownship.Path[ownship.pathIndex]
	var vecToGoal [3]float64
	var stepToGoal [3]float64
	for i := range ownship.position {
		vecToGoal[i] = sub_goal[i] - ownship.position[i]
	}

	goalMagnitude := math.Sqrt((vecToGoal[0] * vecToGoal[0]) + (vecToGoal[1] * vecToGoal[1]) + (vecToGoal[2] * vecToGoal[2]))

	for i := range vecToGoal {
		stepToGoal[i] = (vecToGoal[i] * ownship.stepVelocity) / goalMagnitude
	}

	if ownship.stepVelocity > goalMagnitude {
		ownship.pathIndex += 1
	}
	for i := range stepToGoal {
		ownship.position[i] += stepToGoal[i]
	}
}

type Simulation struct {
	Traffic           Traffic
	Ownship           Ownship
	ConflictDistances [2]float64
	ConflictLog       int
	TimeStep          float64
	T                 int
	conflictRows      []int
}

func (sim *Simulation) Run() {

	for {
		if sim.Ownship.pathIndex >= len(sim.Ownship.Path) {
			sim.End()
			break
		}
		sim.Traffic.Step()
		sim.Ownship.Step()

		for i := 0; i < sim.Traffic.Positions.RawMatrix().Rows; i++ {
			xy_dist := math.Sqrt((sim.Traffic.Positions.At(i, 0)-sim.Ownship.position[0])*(sim.Traffic.Positions.At(i, 0)-sim.Ownship.position[0]) + ((sim.Traffic.Positions.At(i, 1) - sim.Ownship.position[1]) * (sim.Traffic.Positions.At(i, 1) - sim.Ownship.position[1])))
			z_dist := math.Abs(sim.Traffic.Positions.At(i, 2) - sim.Ownship.position[2])
			if xy_dist < sim.ConflictDistances[0] && z_dist < sim.ConflictDistances[1] {
				skip := false
				for r := range sim.conflictRows {
					if i == r {
						skip = true
						break
					}
				}
				if !skip {
					sim.ConflictLog++
					sim.conflictRows = append(sim.conflictRows, i)
				}
			}
		}
		sim.T++
	}

	sim.Traffic.End()
}

func (sim *Simulation) End() {

}
