package s3manager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	awsS3manager "github.com/aws/aws-sdk-go/service/s3/s3manager"

	filesystem "github.com/frozentech/filesystem"
)

// S3Manager ...
type S3Manager struct {
	Config string
	Bucket string
}

// LoadConfig loads the config from env or file
func (a *S3Manager) LoadConfig() error {
	EnviromentVariables := []string{
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"AWS_REGION",
		"AWS_DEFAULT_BUCKET",
		"APP_CONFIG_FOLDER",
	}

	for _, env := range EnviromentVariables {
		if os.Getenv(env) == "" {
			return fmt.Errorf("environment variable %s is missing", env)
		}
	}

	a.Bucket = os.Getenv("AWS_DEFAULT_BUCKET")
	a.Config = os.Getenv("APP_CONFIG_FOLDER")

	return nil
}

// Connect aws
func (a S3Manager) Connect() *session.Session {
	return session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
}

// NewService creates a new s3 service
func (a S3Manager) NewService() *s3.S3 {
	awsSession := a.Connect()
	return s3.New(awsSession)
}

// CreateBucket create bucket
func (a S3Manager) CreateBucket() error {
	if exist, _ := a.BucketExist(); exist {
		return nil
	}

	service := a.NewService()

	_, err := service.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(a.Bucket),
	})

	if err != nil {
		return fmt.Errorf("Unable to create bucket %q, %v", a.Bucket, err)
	}

	err = service.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(a.Bucket),
	})

	if err != nil {
		return fmt.Errorf("Error occurred while waiting for bucket to be created, %s", a.Bucket)
	}

	return nil
}

// CompleteFilename formats the file path
func (a S3Manager) CompleteFilename(filename string) string {
	folder := a.Config
	if len(folder) == 0 {
		folder = "config"
	}
	if folder[len(folder)-1:] == "/" {
		return folder + filename
	}

	return folder + "/" + filename
}

// Delete s3 file
func (a S3Manager) Delete(filename string) (bool, error) {
	localFilename := a.CompleteFilename(filename)
	filesystem.Delete(localFilename)
	service := a.NewService()

	if _, err := service.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(a.Bucket), Key: aws.String(filename)}); err != nil {
		return false, fmt.Errorf("Unable to delete file %s, %v", filename, err)
	}

	return true, nil
}

// List download file
func (a S3Manager) List(bucket string) (objects []*s3.Object, err error) {
	service := a.NewService()
	// Get the list of items
	resp, err := service.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(bucket)})
	if err == nil {
		objects = resp.Contents
	}

	return
}

// Download download file
func (a S3Manager) Download(filename string, args ...string) (bool, error) {
	var localFilename string

	if len(args) > 0 {
		localFilename = args[0]
	} else {
		localFilename = a.CompleteFilename(filename)
		filesystem.Mkdir(filepath.Dir(localFilename))
	}

	file, err := os.Create(localFilename)
	defer file.Close()
	if err != nil {
		return false, fmt.Errorf("Unable to create file %s, %v", localFilename, err)
	}

	session := a.Connect()
	downloader := awsS3manager.NewDownloader(session)

	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(a.Bucket),
			Key:    aws.String(filename),
		})

	if err != nil {
		return false, fmt.Errorf("Unable to download file %s, %v", filename, err)
	}

	return true, nil
}

// Upload uploads file to a s3 bucket
func (a S3Manager) Upload(source string, destination string) (bool, error) {
	session := a.Connect()

	file, err := os.Open(source)
	defer file.Close()
	if err != nil {
		return false, fmt.Errorf("Unable to open file %s, %v", source, err)
	}

	uploader := awsS3manager.NewUploader(session)

	_, err = uploader.Upload(&awsS3manager.UploadInput{
		Bucket: aws.String(a.Bucket),
		Key:    aws.String(destination),
		Body:   file,
	})

	if err != nil {
		return false, fmt.Errorf("Unable to upload %s to %s, %v", source, a.Bucket, err)
	}

	return true, nil
}

// BucketExist checks if bucket exists
func (a S3Manager) BucketExist() (bool, error) {
	service := a.NewService()

	buckets, err := service.ListBuckets(nil)
	if err != nil {
		return false, fmt.Errorf("Unable to list buckets, %v", err)
	}

	for _, b := range buckets.Buckets {
		if aws.StringValue(b.Name) == a.Bucket {
			return true, nil
		}
	}

	return false, fmt.Errorf("Unable to find %s", a.Bucket)
}

// Begin load config and initializes the aws struct
func (a *S3Manager) Begin() error {
	if err := a.LoadConfig(); err != nil {
		return err
	}

	return a.Init()
}

// Init initializes the s3 buckets
func (a *S3Manager) Init() error {
	if err := filesystem.Mkdir(a.Config); err != nil {
		return err
	}

	return a.CreateBucket()
}
