/***
 *  YAML NVMS FORMAT
 *  NVMS Acts as a AWS Parse layer that uses resources from your git repo and information in this config to deploy reproducible apps of any size in efficient microvms
 *  For a typical config we have the following syntax'
 *  To start off we need the base image, you should always use a :minimal variant to get the full benefits of the microvm
 *  Then You'll name the whole project, this is the name of the project that will be used in the AWS Console
 *  Next we'll define services, these can be split by the actual services of your project (e.g. frontend, backend, etc), with different configs for each to allow concur
 *  AWS Service Config
 *  Type, This is the specvific service that we are using
 *  Name, This is the name of the service
 * Engine, This is the engine that we are using
 * Mode: This is the mode that we are using cluster is typically uncommon unless you're building a distributed system
 */