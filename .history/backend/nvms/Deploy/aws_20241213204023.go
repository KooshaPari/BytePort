package deploy

func pushToS3(zipBall []byte, Service lib.Service, AccessKey string, SecretKey string, RootDir string) error {
	// TODO Take Given ZipBall, push Each SubFolder Individually To a singular bucket, keep track of names via map.
}

func provisionEC2