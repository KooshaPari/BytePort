package models

/* A Basic NVMS IAC Config to run say FixIt (Svelte/Gin) we'd need 2 running services for our system to function, actually, ideally 3. Our frontend hosted on an open public port, our backend similarly, and our postgres privately. We want to host all of this on a single microvm instance wherever compute is needed,
and moreover this configuration needs to be such that any other program with the same file structure and build commands could theoretically be deployed on aws via this file, as such this config also needs to directly map to the aws services we need so that each application is fully configured on deploy.

Fixit is a todolist app built on svelte, gin, and a sqlite DB (postgres in byteport)
This is a basic crud app with a minimal ui letting us statically build it, and our backend does not require persistence, which is left to our postgres instance.
As such if we wanted to deploy we would need a firecracker mvm, and a postgres instance alone, with network and security configs declaring our connections between them and end users.

firecracker actions are assumed to be handled by byteport,
(there is the question of when to prefer more mvms when services require more compute or isolation etc, this may be created as an extra spec)

This is very much a WIP configuration

As a result we would need to provide the following in our config, a header with mostly identifying/descriptive info, a declaration of each service with its path build com and port, as well as env vars.
After this your AWSConfig declaring needed services outside of mvms, and the netsec config.


*/
type NVMS struct  {
	Name        string    `yaml:"NAME"`
    Description string    `yaml:"DESCRIPTION"` 
    Services    []Service `yaml:"SERVICES"`
	//AWS      AWSConfig
}
type Service struct {
	Name  string   `yaml:"NAME"`
    Path  string   `yaml:"PATH"`
    Build []string `yaml:"BUILD"` // Changed to slice for command args
    Port  int      `yaml:"PORT"`
    Env   []string `yaml:"ENV,omitempty"`
}
type AWSConfig struct {
	Region   string
	Services []AWSServiceConfig
}
type AWSServiceConfig struct {
	Type       string
	Engine     string
	Mode       string
	Replicas   int
	Size       string
	Name       string
	Partitions int
}
