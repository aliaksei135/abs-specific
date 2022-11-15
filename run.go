package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/aliaksei135/abs-specific/hist"
	"github.com/aliaksei135/abs-specific/sim"
	"github.com/aliaksei135/abs-specific/util"

	"runtime"

	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/urfave/cli/v2"
	// "github.com/urfave/cli/v2/altsrc"
)

func simulateBatch(batch_size int, chan_out chan []int64, bounds [6]float64, alt_hist, track_hist, vel_hist, vert_rate_hist hist.Histogram, timestep, target_density float64, path [][3]float64, conflict_dists [2]float64, surfaceEntrance bool) {
	for i := 0; i < batch_size; i++ {
		seed := rand.Int63()
		traffic := sim.Traffic{Seed: seed, AltitudeDistr: alt_hist, VelocityDistr: vel_hist, TrackDistr: track_hist, VerticalRateDistr: vert_rate_hist, SurfaceEntrance: surfaceEntrance}
		traffic.Setup(bounds, target_density)

		ownVelocity := 60.0
		ownship := sim.Ownship{Path: path, Velocity: ownVelocity}
		ownship.Setup()

		sim := sim.Simulation{Traffic: traffic, Ownship: ownship, ConflictDistances: conflict_dists, TimeStep: timestep}
		sim.Run()
		sim.End()
		pos_sum := 0.0
		samples := int(math.Min(600, float64(len(sim.Traffic.Positions.RawMatrix().Data)-1)))
		for i := 0; i < samples; i++ {
			pos_sum += sim.Traffic.Positions.RawMatrix().Data[i]
		}
		chan_out <- []int64{int64(pos_sum), seed, int64(float64(sim.T) * sim.TimeStep), int64(sim.ConflictLog)}
	}
}

// func parseBounds(boundStr string) [6]float64 {
// 	tokens := strings.Split(boundStr, ",")
// 	var out [6]float64
// 	for i, t := range tokens {
// 		out[i], _ = strconv.ParseFloat(t, 64)
// 	}
// 	return out
// }

// func parseConflictDists(conflictStr string) [2]float64 {
// 	tokens := strings.Split(conflictStr, ",")
// 	var out [2]float64
// 	for i, t := range tokens {
// 		out[i], _ = strconv.ParseFloat(t, 64)
// 	}
// 	return out
// }

func main() {
	log.SetFlags(0)
	start := time.Now()

	app := &cli.App{
		Version:     "0.1a",
		Usage:       "Specific Traffic ABS",
		Description: "Agent Based Traffic MAC Simulation",
		Flags: []cli.Flag{
			&cli.Float64SliceFlag{
				Name:     "bounds",
				Usage:    "W,E,S,N,B,T bounds in metres",
				Required: true,
			},
			&cli.Float64Flag{
				Name:     "target-density",
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
			&cli.PathFlag{
				Name:     "ownPath",
				Usage:    "Path for ownship. Should be a nx3 CSV",
				Required: true,
			},
			&cli.IntFlag{
				Name:  "simOps",
				Usage: "The total number of simulation runs to be done.",
				Value: 1e2,
			},
			&cli.Float64SliceFlag{
				Name:  "conflictDists",
				Usage: "X,Y distances in metres which define a conflict",
				Value: cli.NewFloat64Slice(15.0, 6.0),
			},
			&cli.PathFlag{
				Name:  "dbPath",
				Usage: "A path to the SQLite3 DB the results should be written to",
				Value: "./results.db",
			},
			&cli.Float64Flag{
				Name:  "timestep",
				Usage: "The number of real seconds per simulation timestep. Can be less than 1. Must be greater then 0.",
				Value: 1.0,
			},
			&cli.BoolFlag{
				Name:  "surfaceEntrance",
				Usage: "Boolean flag indicating whether traffic should only spawn at simulation volume surfaces",
				Value: false,
			},
		},
		// Before: altsrc., //TODO Accept file flag input
		Action: func(ctx *cli.Context) error {
			bounds := (*[6]float64)(util.CheckSliceLen(ctx.Float64Slice("bounds"), 6))
			target_density := ctx.Float64("target-density")
			alt_hist := hist.CreateHistogram(util.GetDataFromCSV(util.CheckPathExists(ctx.Path("altDataPath"))), 50)
			track_hist := hist.CreateHistogram(util.GetDataFromCSV(util.CheckPathExists(ctx.Path("trackDataPath"))), 50)
			vel_hist := hist.CreateHistogram(util.GetDataFromCSV(util.CheckPathExists(ctx.Path("velDataPath"))), 50)
			vert_rate_hist := hist.CreateHistogram(util.GetDataFromCSV(util.CheckPathExists(ctx.Path("vertRateDataPath"))), 50)
			own_path := util.GetPathDataFromCSV(util.CheckPathExists(ctx.Path("ownPath")))
			conflict_dist := (*[2]float64)(util.CheckSliceLen(ctx.Float64Slice("conflictDists"), 2))
			dbPath := ctx.Path("dbPath")
			simOps := ctx.Int("simOps")
			timestep := ctx.Float64("timestep")
			surfaceEntrance := ctx.Bool("surfaceEntrance")

			db, err := sql.Open("sqlite3", dbPath)
			if err != nil {
				log.Fatal(err)
			}
			defer db.Close()

			_, err = db.Exec("CREATE TABLE IF NOT EXISTS sims(id, seed, timesteps, n_conflicts)")
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Created/Opened output database")

			result_chan := make(chan []int64)

			n_batches := runtime.NumCPU()
			batch_size := int(simOps / n_batches)
			fmt.Printf("Running %v batches of %v simulations\n", n_batches, batch_size)

			pathLength := util.GetPathLength(own_path)
			expectedSteps := pathLength / 70
			simulatedHours := (expectedSteps * float64(n_batches) * float64(batch_size)) / 3600
			fmt.Printf("Simulating %v hrs, with %v hrs per simulation\n", simulatedHours, expectedSteps/3600)

			for i := 0; i < n_batches; i++ {
				go simulateBatch(batch_size, result_chan, *bounds, alt_hist, track_hist, vel_hist, vert_rate_hist, timestep, target_density, own_path, *conflict_dist, surfaceEntrance)
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
			fmt.Printf("Formatting %v results for database insertion\n", len(sim_results))
			value_fmt := "(%v, %v, %v, %v)"
			string_results := make([]string, len(sim_results))
			for idx, row := range sim_results {
				string_results[idx] = fmt.Sprintf(value_fmt, row[0], row[1], row[2], row[3])
			}
			values_str := strings.Join(string_results, ",")
			fmt.Println("Inserting results into database")
			_, err = db.Exec("INSERT INTO sims VALUES " + values_str)
			if err != nil {
				log.Fatal(err)
				return err
			}

			_, S3Upload := os.LookupEnv("S3_UPLOAD_RESULTS")
			if S3Upload {
				fmt.Println("Uploading results to S3...")
				util.UploadToS3(dbPath)
				fmt.Println("Uploaded results to S3")
			}

			elapsed := time.Since(start).Seconds()
			fmt.Printf("Completed successfully in %v seconds.\n %v ms per simulation.\n %v secs per simulated hour.\n", elapsed, elapsed/float64(1000*n_batches*batch_size), elapsed/simulatedHours)
			fmt.Print("Exiting...\n")
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
