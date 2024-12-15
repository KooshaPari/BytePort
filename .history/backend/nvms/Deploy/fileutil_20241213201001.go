package deploy

func getRootDir (fileMap map[string][]byte) (string, error) {
	for key := range fileMap {
		// Split the key into parts using "/" as the delimiter
		parts := strings.Split(key, "/")
		if len(parts) > 1 {
			// Return the first part (root directory)
			return parts[0] + "/", nil
		}
	}
	return "", fmt.Errorf("no valid root directory found")
}
func processZip (zipBall []byte) (map[string][]byte, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(zipBall), int64(len(zipBall)))
	if err != nil {
		return nil, fmt.Errorf("failed to read zip archive: %v", err)
	}
	fileMap := make(map[string][]byte)
	for _, file := range zipReader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %v", file.Name, err)
		}
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %v", file.Name, err)
		}
		fileMap[file.Name] = content
	}
	return fileMap, nil
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