type ImageComponent struct {
    Name        string
    Version     string
    Platform    string
    Commands    []BuildCommand
}

type BuildCommand struct {
    Type        string      // install, build, configure
    Commands    []string
    EnvVars     map[string]string
}