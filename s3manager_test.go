package s3manager_test

import (
	"encoding/json"
	"os"
	"testing"

	filesystem "github.com/frozentech/filesystem"
	s3manager "github.com/frozentech/s3manager"
)

func TestS3ManagerBegin(t *testing.T) {
	a := s3manager.S3Manager{}
	if err := a.Begin(); err != nil {
		t.Errorf("Expecting `%v`, Got `%v`", err, nil)
	}

	os.Setenv("AWS_REGION", "")
	if err := a.Begin(); err == nil {
		t.Errorf("Expecting `%v`, Got `%v`", err, nil)
	}
	os.Setenv("AWS_REGION", "ap-southeast-1")
}

func TestS3ManagerLoadConfig(t *testing.T) {
	a := s3manager.S3Manager{}
	if err := a.LoadConfig(); err != nil {
		t.Errorf("Expecting `%v`, Got `%v`", err, nil)
	}

	os.Setenv("AWS_REGION", "")
	if err := a.LoadConfig(); err == nil {
		t.Errorf("Expecting `%v`, Got `%v`", err, nil)
	}
	os.Setenv("AWS_REGION", "ap-southeast-1")
}

func TestS3ManagerCreateBucket(t *testing.T) {
	a := s3manager.S3Manager{
		Config: "tester-config",
		Bucket: "tester-bucket",
	}

	exist, _ := a.BucketExist()
	err := a.CreateBucket()

	if exist == false {
		if err == nil {
			t.Errorf("Expecting `%v`, Got `%v`", err, nil)
		}
	} else {
		if err != nil {
			t.Errorf("Expecting `%v`, Got `%v`", nil, err)
		}
	}
}

func TestS3ManagerCompleteFilename(t *testing.T) {
	a := s3manager.S3Manager{
		Config: "tester-config",
		Bucket: "tester-bucket",
	}

	filename := a.CompleteFilename("test-filename")

	if filename != "tester-config/test-filename" {
		t.Errorf("Expecting `%v`, Got `%v`", "tester-config/test-filename", filename)
	}

	a.Config = "tester-config/"

	filename = a.CompleteFilename("test-filename")

	if filename != "tester-config/test-filename" {
		t.Errorf("Expecting `%v`, Got `%v`", "tester-config/test-filename", filename)
	}
}

// JSON converts object to json
func JSON(model interface{}) string {
	body, _ := json.Marshal(model)
	return string(body)
}

func TestS3ManagerUpload(t *testing.T) {
	a := s3manager.S3Manager{}
	a.LoadConfig()

	filename := a.CompleteFilename("test-file")

	filesystem.Mkdir("tester-config")
	filesystem.CreateFile(filename, JSON(a))

	os.Setenv("AWS_REGION", "ap-southeast-1")
	if _, err := a.Upload(filename+".json", "test-file.json"); err != nil {
		t.Errorf("Expecting `%v`, Got `%v`", nil, err)
	}

	if _, err := a.Upload("asdasd.json", "test-file.json"); err == nil {
		t.Errorf("Expecting a error `%v`, Got `%v`", err, nil)
	}

	os.Setenv("AWS_REGION", "ap-southeast-99")
	if _, err := a.Upload(filename+".json", "test-file.json"); err == nil {
		t.Errorf("Expecting a error `%v`, Got `%v`", err, nil)
	}
	os.Setenv("AWS_REGION", "ap-southeast-1")
}

func TestS3ManagerBucketExist(t *testing.T) {
	a := s3manager.S3Manager{
		Bucket: os.Getenv("AWS_DEFAULT_BUCKET"),
		Config: os.Getenv("APP_CONFIG_FOLDER"),
	}
	exist, err := a.BucketExist()
	if !exist {
		t.Errorf("Expecting `%v`, Got `%v`", true, exist)
	}

	if err != nil {
		t.Errorf("Expecting `%v`, Got `%v`", nil, err)
	}

	os.Setenv("AWS_REGION", "ap-southeast-99")
	_, err = a.BucketExist()
	if err == nil {
		t.Errorf("Expecting `%v`, Got `%v`", err, nil)
	}
	os.Setenv("AWS_REGION", "ap-southeast-1")

}

func TestS3ManagerDownload(t *testing.T) {
	a := s3manager.S3Manager{
		Bucket: os.Getenv("AWS_DEFAULT_BUCKET"),
		Config: os.Getenv("APP_CONFIG_FOLDER"),
	}

	if _, err := a.Download("test-file.json"); err != nil {
		t.Errorf("Expecting `%v`, Got `%v` File `%s`", nil, err, "test-file.json")
	}

	if _, err := a.Download("test-file.xml"); err == nil {
		t.Errorf("Expecting `%v`, Got `%v` File `%s`", err, nil, "test-file.xml")
	}

	if _, err := a.Download(".."); err == nil {
		t.Errorf("Expecting `%v`, Got `%v` File `%s`", err, nil, "test-file.xml")
	}

	if _, err := a.Delete("test-file.json"); err != nil {
		t.Errorf("Expecting `%v`, Got `%v` File `%s`", nil, err, "test-file.json")
	}
}

func TestS3ManagerBeginInit(t *testing.T) {
	host := os.Getenv("APP_CONFIG_FOLDER")
	os.Setenv("APP_CONFIG_FOLDER", "")
	defer func() {
		os.Setenv("APP_CONFIG_FOLDER", host)
	}()

	a := s3manager.S3Manager{}
	if err := a.Begin(); err == nil {
		t.Errorf("Expecting `%v`, Got `%v`", err, nil)
	}
}

func TestS3ManagerInit(t *testing.T) {
	a := s3manager.S3Manager{
		Bucket: os.Getenv("AWS_DEFAULT_BUCKET"),
		Config: "test",
	}

	if err := a.Init(); err != nil {
		t.Errorf("Expecting `%v`, Got `%v`", nil, err)
	}
	a.Config = ""

	if err := a.Init(); err == nil {
		t.Errorf("Expecting `%v`, Got `%v`", err, nil)
	}

	a.Config = "tester-config"
	a.Bucket = "tester-config"
	if err := a.Init(); err != nil {
		t.Errorf("Expecting `%v`, Got `%v`", err, nil)
	}

	filesystem.Delete("test")
	filesystem.Delete("tester-config/test-file.json")
	filesystem.Delete("tester-config")
}
