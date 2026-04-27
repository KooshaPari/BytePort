package deploy

func pushToS3(zipBall []byte, Service lib.Service, AccessKeyID) error {
	// TODO Take Given ZipBall, push Each SubFolder Individually To a singular bucket, keep track of names via map.
}