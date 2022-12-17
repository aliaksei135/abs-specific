package main

import (
	_ "net/http/pprof"
	"testing"

	"github.com/aliaksei135/abs-specific/hist"
	"github.com/aliaksei135/abs-specific/util"
	_ "github.com/mattn/go-sqlite3"
)

func Test_simulateBatch(t *testing.T) {
	bounds := [6]float64{-145176.17270300398, -101964.24515822314, 6569893.199178016, 6595219.236650961, 0, 608}
	target_density := 1e-10
	alt_hist := hist.CreateHistogram(util.GetDataFromCSV(util.CheckPathExists("test_data/alts.csv")), 50)
	track_hist := hist.CreateHistogram(util.GetDataFromCSV(util.CheckPathExists("test_data/tracks.csv")), 50)
	vel_hist := hist.CreateHistogram(util.GetDataFromCSV(util.CheckPathExists("test_data/vels.csv")), 50)
	vert_rate_hist := hist.CreateHistogram(util.GetDataFromCSV(util.CheckPathExists("test_data/vert_rates.csv")), 50)
	own_path := util.GetPathDataFromCSV(util.CheckPathExists("test_data/path.csv"))
	own_velocity := 60.0
	conflict_dist := [2]float64{15, 6}
	timestep := 0.1
	surfaceEntrance := false

	batch_size := 200

	result_chan := make(chan []int64, batch_size+1)
	defer close(result_chan)

	simulateBatch(batch_size, 0, result_chan, bounds, alt_hist, track_hist, vel_hist, vert_rate_hist, timestep, target_density, own_velocity, own_path, conflict_dist, surfaceEntrance)
}
