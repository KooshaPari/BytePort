package lib

func matchesBuildpackInMemory(tree map[string][]string, detectFiles []string, servicePath string, rootDir string) bool {
    normalizedPath := filepath.Base(strings.Trim(servicePath, "/"))  // Get just 'backend' or 'frontend'
    //fmt.Printf("Checking buildpacks for path: %s\n", normalizedPath)
    //fmt.Printf("Files to detect: %v\n", detectFiles)
    //fmt.Printf("Tree: %+v\n", tree)

    // Get files in the service directory
    dirFiles, exists := tree[normalizedPath]
    if !exists {
        fmt.Printf("Directory %s not found in tree\n", normalizedPath)
        return false
    }

    //fmt.Printf("Files in %s: %v\n", normalizedPath, dirFiles)
    
    // Check each detect file
    for _, file := range detectFiles {
        for _, f := range dirFiles {
            if f == file {
                fmt.Printf("Found matching file %s in %s\n", file, normalizedPath)
                return true
            }
        }
    }

    fmt.Printf("No matching files found in %s\n", normalizedPath)
    return false
}
 func AnalyzeBuildpackPaths(paths []string) (map[string ][]string,string, error) {
    // Extract root directory
    rootDir := findCommonPrefix(paths)
    if rootDir == "" {
        return nil, "", fmt.Errorf("no common root directory found")
    }

    // Create file tree
    tree := make(map[string][]string)
    for _, path := range paths {
        relativePath := strings.TrimPrefix(path, rootDir)
        dir := filepath.Dir(relativePath)
        tree[dir] = append(tree[dir], filepath.Base(relativePath))
    }

       return tree, rootDir,nil
    }

   
func findCommonPrefix(paths []string) string {
    if len(paths) == 0 {
        return ""
    }

    prefix := paths[0]
    for _, path := range paths[1:] {
        for !strings.HasPrefix(path, prefix) {
            prefix = prefix[:strings.LastIndex(prefix, "/")]
        }
    }
    return prefix
}