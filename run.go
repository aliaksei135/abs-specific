package main

import (
	"abs-specific/hist"
	"abs-specific/sim"
	"database/sql"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"gonum.org/v1/gonum/mat"
)

func simulateBatch(batch_size int, chan_out chan []int64, bounds [6]float64, alt_hist, track_hist, vel_hist, vert_rate_hist hist.Histogram, target_density float64, path []mat.Dense, conflict_dists [2]float64) {

	for i := 0; i < batch_size; i++ {
		seed := rand.Int63()
		traffic := sim.Traffic{Seed: seed, AltitudeDistr: alt_hist, VelocityDistr: vel_hist, TrackDistr: track_hist, VerticalRateDistr: vert_rate_hist}
		traffic.Setup(bounds, target_density)

		ownship := sim.Ownship{Path: path, Velocity: 10.0}
		ownship.Setup()

		sim := sim.Simulation{Traffic: traffic, Ownship: ownship, ConflictDistances: conflict_dists}
		sim.Run()
		sim.End()
		pos_sum := 0.0
		samples := int(math.Max(600, float64(len(sim.Traffic.Positions.RawMatrix().Data))))
		for i := 0; i < samples; i++ {
			pos_sum += sim.Traffic.Positions.RawMatrix().Data[i]
		}
		chan_out <- []int64{int64(pos_sum), seed, int64(sim.T), int64(sim.ConflictLog)}
	}
}

func GetPathDataFromCSV(csvPath string) []mat.Dense {
	file, err := os.Open(csvPath)
	if err != nil {
		fmt.Println(err)
	}
	reader := csv.NewReader(file)
	vals, _ := reader.ReadAll()
	out := make([]mat.Dense, len(vals))
	for i, str := range vals {
		x, _ := strconv.ParseFloat(str[0], 64)
		y, _ := strconv.ParseFloat(str[1], 64)
		z, _ := strconv.ParseFloat(str[2], 64)
		out[i] = *mat.NewDense(1, 3, []float64{x, y, z})
	}
	return out
}

func parseBounds(boundStr string) [6]float64 {
	tokens := strings.Split(boundStr, ",")
	var out [6]float64
	for i, t := range tokens {
		out[i], _ = strconv.ParseFloat(t, 64)
	}
	return out
}

func parseConflictDists(conflictStr string) [2]float64 {
	tokens := strings.Split(conflictStr, ",")
	var out [2]float64
	for i, t := range tokens {
		out[i], _ = strconv.ParseFloat(t, 64)
	}
	return out
}

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.Float64SliceFlag{
				Name:     "bounds",
				Usage:    "S,W,N,E bounds in metres",
				Required: true,
			},
			&cli.Float64Flag{
				Name:     "targetDensity",
				Usage:    "Target background traffic density in ac/m^3",
				Required: true,
			},
			&cli.PathFlag{
				Name:     "altDataPath",
				Usage:    "Path to altitude data in metres as CSV",
				Required: true,
			},
			&cli.PathFlag{
				Name:     "velDataPath",
				Usage:    "Path to velocity data in m/s as CSV",
				Required: true,
			},
			&cli.PathFlag{
				Name:     "trackDataPath",
				Usage:    "Path to track data in deg as CSV",
				Required: true,
			},
			&cli.PathFlag{
				Name:     "vertRateDataPath",
				Usage:    "Path to vertical rate data in m/s as CSV",
				Required: true,
			},
			&cli.Float64SliceFlag{
				Name:     "ownPath",
				Usage:    "Path for ownship. Should be a nx3 CSV",
				Required: true,
			},
			&cli.IntFlag{
				Name:  "simOps",
				Usage: "The total number of simluation runs to be done.",
				Value: 1e9,
			},
			&cli.Float64SliceFlag{
				Name:  "conflictDists",
				Usage: "X,Y distances in metres which define a conflict",
			},
		},
		Action: func(cCtx *cli.Context) error {
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

	boundsArg := flag.String("bounds", "", "S,W,N,E bounds in metres")
	targetDensityArg := flag.Float64("target-density", 1e-9, "Target Background Traffic Density")
	altsData := flag.String("alts-data", "", "Path to altitude data as CSV")
	velsData := flag.String("vels-data", "", "Path to velocity data as CSV")
	tracksData := flag.String("tracks-data", "", "Path to track data as CSV")
	vertRatesData := flag.String("vert-rates-data", "", "Path to vertical rate data as CSV")
	pathData := flag.String("path-data", "", "Path to ownship trajectory as CSV")
	dbPath := flag.String("output-path", "", "Path to results database")
	simOps := flag.Int("sim-ops", 1e3, "Number of operations to simulate")
	conflictDists := flag.String("conflict-dists", "20,20", "X,Y distances in metres which define a conflict")

	flag.Parse()

	bounds := parseBounds(*boundsArg)
	target_density := *targetDensityArg
	alt_hist := hist.CreateHistogram(hist.GetDataFromCSV(*altsData), 50)
	track_hist := hist.CreateHistogram(hist.GetDataFromCSV(*tracksData), 50)
	vel_hist := hist.CreateHistogram(hist.GetDataFromCSV(*velsData), 50)
	vert_rate_hist := hist.CreateHistogram(hist.GetDataFromCSV(*vertRatesData), 50)
	own_path := GetPathDataFromCSV(*pathData)
	conflict_dist := parseConflictDists(*conflictDists)

	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS sims(id, seed, timesteps, n_conflicts)")
	if err != nil {
		log.Fatal(err)
	}

	result_chan := make(chan []int64)

	n_batches := runtime.NumCPU()
	batch_size := int(*simOps / n_batches)

	for i := 0; i < n_batches; i++ {
		go simulateBatch(batch_size, result_chan, bounds, alt_hist, track_hist, vel_hist, vert_rate_hist, target_density, own_path, conflict_dist)
	}

	sim_results := make([][]int64, n_batches*batch_size)

	result_count := 0
	for results := range result_chan {
		sim_results[result_count] = results

		result_count++
		if result_count >= n_batches*batch_size {
			break
		}
	}

	value_fmt := "(%v, %v, %v, %v)"
	string_results := make([]string, len(sim_results))
	for idx, row := range sim_results {
		string_results[idx] = fmt.Sprintf(value_fmt, row[0], row[1], row[2], row[3])
	}
	values_str := strings.Join(string_results, ",")
	_, err = db.Exec("INSERT INTO sims VALUES " + values_str)
	if err != nil {
		log.Fatal(err)
	}
}
