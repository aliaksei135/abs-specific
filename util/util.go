package util

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gonum.org/v1/gonum/mat"
)

func GetDataFromCSV(csvPath string) []float64 {
	file, err := os.Open(csvPath)
	if err != nil {
		log.Fatal(err)
	}
	reader := csv.NewReader(file)
	vals, _ := reader.ReadAll()
	out := make([]float64, len(vals))
	for i, str := range vals {
		out[i], _ = strconv.ParseFloat(str[0], 64)
	}
	return out
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
		x, _ := strconv.ParseFloat(strings.TrimPrefix(str[0], "\uFEFF"), 64)
		y, _ := strconv.ParseFloat(strings.TrimPrefix(str[1], "\uFEFF"), 64)
		z, _ := strconv.ParseFloat(strings.TrimPrefix(str[2], "\uFEFF"), 64)
		out[i] = *mat.NewDense(1, 3, []float64{x, y, z})
	}
	return out
}

func GetPathLength(path []mat.Dense) float64 {
	length := 0.0
	for i := 0; i < len(path)-1; i++ {
		var diff mat.Dense
		diff.Sub(&path[i], &path[i+1])
		length += diff.Norm(2)
	}
	return length
}

func CheckPathExists(path string) string {
	_, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	return path
}

func CheckSliceLen[T any](slice []T, requiredLength int) []T {
	if len(slice) != requiredLength {
		log.Fatalf("Incorrect slice length. Wanted length %v, got %v for slice %v", requiredLength, len(slice), slice)
	}
	return slice
}
