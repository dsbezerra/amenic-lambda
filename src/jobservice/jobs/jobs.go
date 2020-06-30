package jobs

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence"
	"github.com/dsbezerra/amenic-lambda/src/lib/persistence/mongolayer"
	"github.com/pkg/errors"
)

// Input ...
type Input struct {
	Name string   `json:"name"`
	Args []string `json:"args"`
}

// Handler ...
type Handler func(in *Input, data persistence.DataAccessLayer) error

const (
	// DefaultQueueName is the default queue used to processs all amenic jobs
	DefaultQueueName = "AmenicJobQueue"

	// Job ...
	Job = "Job"

	// JobCreateStatic indicates the desired job to run is create_static
	JobCreateStatic = "create_static"
	// JobCheckOpeningMovies indicates the desired job to run is check_opening_movies
	JobCheckOpeningMovies = "check_opening_movies"
	// JobStartScraper indicates the desired job to run is start_scraper
	JobStartScraper = "start_scraper"
	// JobSyncScores indicates the desired job to run is sync_scores
	JobSyncScores = "sync_scores"
	// JobUploadImagesToCloudinary indicates the desired job to run is upload_images_to_cloudinary
	JobUploadImagesToCloudinary = "upload_images_to_cloudinary"
)

var (
	// Handlers ...
	Handlers = map[string]Handler{
		JobCreateStatic:             CreateStatic,
		JobCheckOpeningMovies:       CheckOpeningMovies,
		JobStartScraper:             StartScrapers,
		JobSyncScores:               SyncScores,
		JobUploadImagesToCloudinary: UploadImagesToCloudinary,
	}
)

// SendMessageToQueue ...
func SendMessageToQueue(input *sqs.SendMessageInput) (string, error) {
	if input == nil {
		return "", errors.New("invalid send message input")
	}

	sess, err := session.NewSession(&aws.Config{Region: aws.String("sa-east-1")})
	svc := sqs.New(sess)

	// Check if we have the desired queue created
	queues, err := svc.ListQueues(nil)
	if err != nil {
		return "", err
	}

	var queue *string
	for _, url := range queues.QueueUrls {
		if url == nil {
			continue
		}
		if strings.Contains(*url, DefaultQueueName) {
			queue = url
			break
		}
	}

	if queue == nil {
		result, err := svc.CreateQueue(&sqs.CreateQueueInput{
			QueueName: aws.String(DefaultQueueName),
			Attributes: map[string]*string{
				"MessageRetentionPeriod": aws.String("86400"),
			},
		})
		if err != nil {
			return "", err
		}
		queue = result.QueueUrl
	}

	if queue == nil {
		return "", errors.New("invalid queue")
	}

	// Actually send message
	input.SetQueueUrl(*queue)
	result, err := svc.SendMessage(input)
	if err != nil {
		return "", err
	}
	return *result.MessageId, nil
}

// UploadToBucket ...
func UploadToBucket(bucket string, filename string, file []byte) error {
	sess, err := session.NewSession(&aws.Config{Region: aws.String("sa-east-1")})
	if err != nil {
		return err
	}

	uploader := s3manager.NewUploader(sess)

	fmt.Printf("Uploading %s to bucket %s...", filename, bucket)

	body := bytes.NewReader(file)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
		Body:   body,
	})

	if err != nil && err.Error() == s3.ErrCodeNoSuchBucket {
		svc := s3.New(sess)
		_, err = svc.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			return err
		}
		err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{Bucket: aws.String(bucket)})
		if err != nil {
			return err
		}
		return UploadToBucket(bucket, filename, file)
	}

	if err == nil {
		fmt.Printf(" Success!")
	}
	fmt.Println()
	return err
}

// DeleteFromBucket ...
func DeleteFromBucket(bucket string, filename string) error {
	sess, err := session.NewSession(&aws.Config{Region: aws.String("sa-east-1")})
	if err != nil {
		return err
	}
	svc := s3.New(sess)
	_, err = svc.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(bucket), Key: aws.String(filename)})
	return err
}

// GetDataAccessLayer ...
func GetDataAccessLayer() (persistence.DataAccessLayer, error) {
	var envName string
	if os.Getenv("AMENIC_MODE") == "debug" {
		envName = "DB_DEV"
	} else {
		envName = "DB_PROD"
	}
	data, err := mongolayer.NewMongoDAL(os.Getenv(envName))
	if err != nil {
		return nil, err
	}
	// data.Setup()
	return data, nil
}

func parseArgs(in *Input) map[string]string {
	l := len(in.Args)
	if l == 0 || l%2 != 0 {
		return nil
	}

	var i int

	r := map[string]string{}
	for i < l {
		v := in.Args[i]
		if !strings.HasPrefix(v, "-") {
			// Ignore the rest
			break
		} else {
			r[v[1:]] = in.Args[i+1]
			i += 2
		}
	}

	return r
}
