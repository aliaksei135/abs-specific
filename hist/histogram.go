package hist

import (
	"math/rand"
	"sort"
)

type Histogram struct {
	bin_midpoints []float64
	cdf           []float64
}

func CreateHistogram(data []float64, num_bins int) Histogram {
	sort.Float64s(data)
	data_min, data_max := data[0], data[len(data)-1]

	cdf := make([]float64, num_bins)
	hist := make([]int, num_bins)
	bin_edges := make([]float64, num_bins+1)
	bin_midpoints := make([]float64, num_bins)

	//Find bin edges
	interval := (data_max - data_min) / float64(num_bins)
	for i := 0; i < num_bins+1; i++ {
		bin_edges[i] = data_min + (float64(i) * interval)
	}
	//Find bin midpoints and populate histogram
	for i := 0; i < num_bins; i++ {
		left_edge := bin_edges[i]
		right_edge := bin_edges[i+1]
		bin_midpoints[i] = (left_edge + right_edge) / 2
		left_idx := sort.SearchFloat64s(data, left_edge)
		right_idx := sort.SearchFloat64s(data, right_edge)
		hist[i] = right_idx - left_idx
	}

	var cumsum float64
	for i, val := range hist {
		cumsum += float64(val)
		cdf[i] = cumsum
	}
	// Normalise
	for i := range cdf {
		cdf[i] = cdf[i] / cumsum
	}
	return Histogram{bin_midpoints: bin_midpoints, cdf: cdf}
}

func (hist *Histogram) Sample(num int) []float64 {
	samples := make([]float64, num)
	for i := 0; i < num; i++ {
		randn := rand.Float64()
		insert_idx := sort.SearchFloat64s(hist.cdf, randn)
		samples[i] = hist.bin_midpoints[insert_idx]
	}
	return samples
}
