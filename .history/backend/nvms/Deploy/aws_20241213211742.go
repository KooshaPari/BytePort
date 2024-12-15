package deploy

import (
	"archive/zip"
	"bytes"
	"fmt"
	"nvms/lib"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func pushToS3(zipBall []byte, AccessKey string, SecretKey string, ProjectName string) error {
	// TODO Take Given ZipBall, push Each SubFolder Individually To a singular bucket, keep track of names via map.
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Region:      aws.String("us-east-1"),
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	svc := s3.New(sess)
	bucketName := "NVMS-"+ProjectName
	// Create the bucket
	createBucketInput := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}
	if _, err := svc.CreateBucket(createBucketInput); err != nil {
		return fmt.Errorf("failed to create bucket: %v", err)
	}
	// Parse the zip archive
	zipReader, err := zip.NewReader(bytes.NewReader(zipBall), int64(len(zipBall)))
	if err != nil {
		return fmt.Errorf("failed to read zip archive: %v", err)
	}
	// Upload each file in the zip archive to the bucket
	for i, file := range zipReader.File {
		fileReader, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip archive: %v", err)
		}
		defer fileReader.Close()
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(fileReader); err != nil {
			return fmt.Errorf("failed to read file content: %v", err)
		}
		uploadInput := &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(file.Name),
			Body:   bytes.NewReader(buf.Bytes()),
		}
		_, err = svc.PutObject(uploadInput)
		if err != nil {
			return fmt.Errorf("failed to upload file %d to bucket: %v", i, err)
		}
	}
	return nil
}

func provisionEC2(zipBall []byte, Service lib.Service, AccessKey string, SecretKey string, RootDir string) error {
}

func getServiceFolder(name string, zipBall []byte) (map[string][]byte,string, error) {
	fileMap, err := processZip(zipBall)
	if err != nil {
		return nil, "", fmt.Errorf("failed to process zip file: %v", err)
	}
	rootDir, err := getRootDir(fileMap)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get root directory: %v", err)
	}

	serviceDir := rootDir + name
	if _, ok := fileMap[serviceDir]; !ok {
		return nil, "", fmt.Errorf("service directory %s not found in zip file", serviceDir)
	}
	return fileMap, serviceDir, nil
}
