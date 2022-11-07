package sim

import (
	"abs-specific/hist"
	"testing"

	"gonum.org/v1/gonum/mat"
)

func Test_bearing2angle(t *testing.T) {
	type args struct {
		bearing float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{"N-0", args{0.0}, 90},
		{"N-360", args{360.0}, 90},
		{"E", args{90.0}, 0},
		{"S", args{180.0}, 270},
		{"W", args{270.0}, 180},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bearing2angle(tt.args.bearing); got != tt.want {
				t.Errorf("bearing2angle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTraffic_Setup(t *testing.T) {
	alt_hist := hist.CreateHistogram(hist.GetDataFromCSV("../test_data/alts.csv"), 20)
	track_hist := hist.CreateHistogram(hist.GetDataFromCSV("../test_data/tracks.csv"), 20)
	vel_hist := hist.CreateHistogram(hist.GetDataFromCSV("../test_data/vels.csv"), 20)
	vert_rate_hist := hist.CreateHistogram(hist.GetDataFromCSV("../test_data/vert_rates.csv"), 20)

	type args struct {
		bounds         [6]float64
		target_density float64
	}
	tests := []struct {
		name string
		tfc  *Traffic
		args args
	}{
		{"Setup", &Traffic{Seed: 321, AltitudeDistr: alt_hist, VelocityDistr: vel_hist, TrackDistr: track_hist, VerticalRateDistr: vert_rate_hist}, args{[6]float64{0, 1e4, 0, 1e4, 0, 1524}, 4e-9}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tfc.Setup(tt.args.bounds, tt.args.target_density)
		})
	}
}

func TestTraffic_Step(t *testing.T) {
	alt_hist := hist.CreateHistogram(hist.GetDataFromCSV("../test_data/alts.csv"), 40)
	track_hist := hist.CreateHistogram(hist.GetDataFromCSV("../test_data/tracks.csv"), 40)
	vel_hist := hist.CreateHistogram(hist.GetDataFromCSV("../test_data/vels.csv"), 40)
	vert_rate_hist := hist.CreateHistogram(hist.GetDataFromCSV("../test_data/vert_rates.csv"), 40)
	traffic := Traffic{Seed: 321, AltitudeDistr: alt_hist, VelocityDistr: vel_hist, TrackDistr: track_hist, VerticalRateDistr: vert_rate_hist}
	traffic.Setup([6]float64{0, 1e4, 0, 1e4, 0, 1524}, 4e-9)
	tests := []struct {
		name string
		tfc  *Traffic
	}{
		{"Step", &traffic},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tfc.Step()
		})
	}
}

func TestOwnship_Step(t *testing.T) {
	path := []mat.Dense{*mat.NewDense(1, 3, []float64{1, 1, 200}), *mat.NewDense(1, 3, []float64{300, 600, 800}), *mat.NewDense(1, 3, []float64{2000, 5000, 900}), *mat.NewDense(1, 3, []float64{3000, 6000, 200})}
	ownship := Ownship{Path: path, Velocity: 10.0}
	ownship.Setup()

	tests := []struct {
		name    string
		ownship *Ownship
	}{
		{"Step", &ownship},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ownship.Step()
		})
	}
}

func TestSimulation_Run(t *testing.T) {
	alt_hist := hist.CreateHistogram(hist.GetDataFromCSV("../test_data/alts.csv"), 40)
	track_hist := hist.CreateHistogram(hist.GetDataFromCSV("../test_data/tracks.csv"), 40)
	vel_hist := hist.CreateHistogram(hist.GetDataFromCSV("../test_data/vels.csv"), 40)
	vert_rate_hist := hist.CreateHistogram(hist.GetDataFromCSV("../test_data/vert_rates.csv"), 40)
	traffic := Traffic{Seed: 321, AltitudeDistr: alt_hist, VelocityDistr: vel_hist, TrackDistr: track_hist, VerticalRateDistr: vert_rate_hist}
	traffic.Setup([6]float64{0, 1e4, 0, 1e4, 0, 1524}, 1e-7)

	path := []mat.Dense{*mat.NewDense(1, 3, []float64{1, 1, 200}), *mat.NewDense(1, 3, []float64{300, 600, 800}), *mat.NewDense(1, 3, []float64{2000, 5000, 900}), *mat.NewDense(1, 3, []float64{3000, 6000, 200})}
	ownship := Ownship{Path: path, Velocity: 10.0}
	ownship.Setup()

	sim := Simulation{Traffic: traffic, Ownship: ownship, ConflictDistances: [2]float64{20, 20}}

	tests := []struct {
		name string
		sim  *Simulation
	}{
		{"Run Sim", &sim},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.sim.Run()
		})
	}
}
