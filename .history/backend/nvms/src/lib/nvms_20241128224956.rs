/***
 *  YAML NVMS FORMAT
 *  NVMS Acts as a AWS Parse layer that uses resources from your git repo and information in this config to deploy reproducible apps of any size in efficient microvms
 *  For a typical config we have the following syntax'
 *  To start off we need the base image, you should always use a :minimal variant to get the full benefits of the microvm
 *  Then You'll name the whole project, this is the name of the project that will be used in the AWS Console
 *  Next we'll define services, these can be split by the actual services of your project (e.g. frontend, backend, etc), with different configs for each to allow concurrent deployment
 *  Each one requires a path to the service, a build script / command, a port, and environment variables (optional)
 *  We also have present scalability rules, (Min, MAx, CPU Threshold, Memory Threshold) that will allow you to set a range of resources and a threshold to increment(decr = thresh/2)
 *  If you have a distributed system you'll need to additionally configur a cluster, this will allow you to deploy multiple instances of the same service on different microvms
 *  furtherore we can directly specify our entire aws infra through the SERVICES: section incl MODE and Engine (e.g. ECS, EKS, etc) to allow for a more complex deployment, 
 *  Network options while mostly self contained are also present to allow for more complex networking options
 *  AWS Service Config
 *  Type, This is the specvific service that we are using
 *  Name, This is the name of the service
 * Engine, This is the engine that we are using
 * Mode: This is the mode that we are using cluster is typically uncommon unless you're building a distributed system
 */

pub struct NVMS {
    pub FROM: String,
    pub NAME: Option<String>,
    pub SERVICES: Option<Vec<Service>>,
    pub RESOURCES: Option<Vec<Resource>>,
    pub CLUSTER: Option<Vec<Cluster>>,
    pub AWS: Option<Vec<AWS>>,
    pub NETWORK: Option<Vec<Network>>,
    pub SCALE: Option<Vec<Scale>>,


}
pub struct Service {
    pub PATH: String,
    pub BUILD: String,
    pub PORT: String,
    pub ENV: Option<Vec<String>>,
}
pub struct Resource {
    pub MODE: String,
    pub INSTANCES: vec<Instance>,
}
pub struct Instance {
    pub NAME: String,
    pub CPU: String,
    pub MEM: String,
    pub ENV: Option<Vec<String>>,
    pub MEM: String,
}

pub struct AWS {
    pub REGION: String,
    pub SERVICES: vec<AWSservice>,
}
pub struct AWSservice {
    pub TYPE: String,
    pub NAME: String,
    pub ENGINE: String,
    pub MODE: String,
}
pub struct Network {
    pub DOMAIN: String,
    
}
pub struct Cluster {
    pub NAME: String,
    pub COUNT: String,
}
pub struct Scale {
    pub MIN: String,
    pub MAX: String,
    pub CPU: String,
    pub MEM: String,
}
