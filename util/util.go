package util

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"gonum.org/v1/gonum/graph/set/uid"
)

var (
	S3SetupComplete bool
	S3Downloader    s3manager.Downloader
	S3Uploader      s3manager.Uploader
)

func setupS3() {
	S3_KEY := os.Getenv("S3_KEY")
	S3_SECRET := os.Getenv("S3_SECRET")
	S3_REGION := os.Getenv("S3_REGION")
	if S3_KEY == "" {
		panic("S3 key empty")
	}
	if S3_SECRET == "" {
		panic("S3 secret empty")
	}
	if S3_REGION == "" {
		panic("S3 region empty")
	}

	sess, _ := session.NewSession(&aws.Config{
		Region:      aws.String(S3_REGION),
		Credentials: credentials.NewStaticCredentials(S3_KEY, S3_SECRET, ""),
	},
	)

	S3Downloader = *s3manager.NewDownloader(sess)
	S3Uploader = *s3manager.NewUploader(sess)
	S3SetupComplete = true
}

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

func GetPathDataFromCSV(csvPath string) [][3]float64 {
	file, err := os.Open(csvPath)
	if err != nil {
		fmt.Println(err)
	}
	reader := csv.NewReader(file)
	vals, _ := reader.ReadAll()
	out := make([][3]float64, len(vals))
	for i, str := range vals {
		x, _ := strconv.ParseFloat(strings.TrimPrefix(str[0], "\uFEFF"), 64)
		y, _ := strconv.ParseFloat(strings.TrimPrefix(str[1], "\uFEFF"), 64)
		z, _ := strconv.ParseFloat(strings.TrimPrefix(str[2], "\uFEFF"), 64)
		out[i] = [3]float64{x, y, z}
	}
	return out
}

func GetPathLength(path [][3]float64) float64 {
	length := 0.0
	for i := 0; i < len(path)-1; i++ {
		currentPoint := path[i]
		nextPoint := path[i+1]
		dist := math.Sqrt((currentPoint[0]-nextPoint[0])*(currentPoint[0]-nextPoint[0]) + ((currentPoint[1] - nextPoint[1]) * (currentPoint[1] - nextPoint[1])) + ((currentPoint[2] - nextPoint[2]) * (currentPoint[2] - nextPoint[2])))
		length += dist
	}
	return length
}

func CheckPathExists(path string) string {
	if strings.HasPrefix(strings.ToLower(path), "s3://") {
		if !S3SetupComplete {
			setupS3()
		}
		tokens := strings.Split(path, "/")
		keyName := tokens[len(tokens)-1]
		bucketName := tokens[len(tokens)-2]

		filePath := os.TempDir() + fmt.Sprint(os.PathSeparator) + fmt.Sprint(uid.NewSet().NewID()) + keyName
		file, err := os.Open(filePath)
		if err != nil {
			panic("Could not create temp file to download S3 file " + path)
		}

		_, err = S3Downloader.Download(file,
			&s3.GetObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(keyName),
			},
		)
		if err != nil {
			panic("Could not download from S3")
		}

		return filePath

	} else {

		_, err := os.Open(path)
		if err != nil {
			log.Fatal(err)
		}
		return path
	}
}

func CheckSliceLen[T any](slice []T, requiredLength int) []T {
	if len(slice) != requiredLength {
		log.Fatalf("Incorrect slice length. Wanted length %v, got %v for slice %v", requiredLength, len(slice), slice)
	}
	return slice
}
