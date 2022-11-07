package util

import (
	"reflect"
	"testing"
)

// func TestGetDataFromCSV(t *testing.T) {
// 	type args struct {
// 		csvPath string
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want []float64
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := GetDataFromCSV(tt.args.csvPath); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("GetDataFromCSV() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func TestGetPathDataFromCSV(t *testing.T) {
	type args struct {
		csvPath string
	}
	tests := []struct {
		name string
		args args
		want [][3]float64
	}{
		{"pathImport", args{"../test_data/path.csv"}, [][3]float64{{-119012.4530400194053072482347, 6594719.2708460101857781410217, 1000.0000000000000000000000}, {-112477.0871216368541354313493, 6594642.8338177241384983062744, 1000.0000000000000000000000}, {-104680.5102365488564828410745, 6592731.9081105943769216537476, 1000.0000000000000000000000}, {-102463.8364162787620443850756, 6588451.4345266232267022132874, 1000.0000000000000000000000}, {-104107.2325244100502459332347, 6581954.2871223827823996543884, 1000.0000000000000000000000}, {-136363.6584607544355094432831, 6575724.6693171421065926551819, 1000.0000000000000000000000}, {-143548.7391195610107388347387, 6570985.5735634621232748031616, 1000.0000000000000000000000}, {-144676.1852867673442233353853, 6570393.1865942524746060371399, 1000.0000000000000000000000}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetPathDataFromCSV(tt.args.csvPath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPathDataFromCSV() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPathLength(t *testing.T) {
	type args struct {
		path [][3]float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{"pathLength", args{GetPathDataFromCSV("../test_data/path.csv")}, 68818.61266545641},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetPathLength(tt.args.path); got != tt.want {
				t.Errorf("GetPathLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckPathExists(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"existingPath", args{"../test_data/path.csv"}, "../test_data/path.csv"},
		{"existingPath", args{"../test_data/alts.csv"}, "../test_data/alts.csv"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckPathExists(tt.args.path); got != tt.want {
				t.Errorf("CheckPathExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestCheckSliceLen(t *testing.T) {
// 	type args struct {
// 		slice          []T
// 		requiredLength int
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want []T
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := CheckSliceLen(tt.args.slice, tt.args.requiredLength); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("CheckSliceLen() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
