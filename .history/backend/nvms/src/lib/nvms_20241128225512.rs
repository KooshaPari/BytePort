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
use std::collections::HashMap;
use yaml2::Yaml;
use thiserror::Error;

// First, let's define our error types
#[derive(Error, Debug)]
pub enum NVMSError {
    #[error("YAML parsing error: {0}")]
    YamlError(String),
    #[error("Missing required field: {0}")]
    MissingField(String),
    #[error("Invalid value for field {field}: {message}")]
    InvalidValue {
        field: String,
        message: String,
    },
}

// Core types for common patterns
#[derive(Debug, Clone)]
pub struct ResourceSize {
    pub size: u32,
    pub unit: ResourceUnit,
}

#[derive(Debug, Clone)]
pub enum ResourceUnit {
    GB,
    MB,
    KB,
}

// Main NVMS Configuration
#[derive(Debug)]
pub struct NVMSConfig {
    pub from: String,
    pub name: String,
    pub services: Option<HashMap<String, ServiceConfig>>,
    pub cluster: Option<HashMap<String, ClusterConfig>>,
    pub aws: Option<AWSConfig>,
    pub network: Option<NetworkConfig>,
}

impl TryFrom<Yaml> for NVMSConfig {
    type Error = NVMSError;

    fn try_from(yaml: Yaml) -> Result<Self, Self::Error> {
        let map = yaml.as_hash().ok_or(NVMSError::YamlError(
            "Root must be a mapping".to_string(),
        ))?;

        Ok(Self {
            from: map.get_str("FROM")?.to_string(),
            name: map.get_str("NAME")?.to_string(),
            services: map.get_opt("SERVICES")?.map(ServiceConfig::try_from_yaml),
            cluster: map.get_opt("CLUSTER")?.map(ClusterConfig::try_from_yaml),
            aws: map.get_opt("AWS")?.map(AWSConfig::try_from_yaml),
            network: map.get_opt("NETWORK")?.map(NetworkConfig::try_from_yaml),
        })
    }
}

// Service Configuration
#[derive(Debug)]
pub struct ServiceConfig {
    pub path: String,
    pub build: BuildCommand,
    pub port: Option<u16>,
    pub env: Option<HashMap<String, String>>,
    pub resources: Option<ResourceConfig>,
}

#[derive(Debug)]
pub enum BuildCommand {
    Single(String),
    Multiple(Vec<String>),
}

impl ServiceConfig {
    fn try_from_yaml(yaml: &Yaml) -> Result<Self, NVMSError> {
        let map = yaml.as_hash().ok_or(NVMSError::YamlError(
            "Service must be a mapping".to_string(),
        ))?;

        let build = match map.get("BUILD") {
            Some(b) if b.is_string() => BuildCommand::Single(b.as_str().unwrap().to_string()),
            Some(b) if b.is_array() => BuildCommand::Multiple(
                b.as_vec()
                    .unwrap()
                    .iter()
                    .map(|s| s.as_str().unwrap().to_string())
                    .collect(),
            ),
            _ => return Err(NVMSError::MissingField("BUILD".to_string())),
        };

        Ok(Self {
            path: map.get_str("PATH")?.to_string(),
            build,
            port: map.get_opt("PORT")?.map(|p| p.as_u16().unwrap()),
            env: map.get_opt("ENV")?.map(parse_env_vars),
            resources: map.get_opt("RESOURCES")?.map(ResourceConfig::try_from_yaml),
        })
    }
}

// Helper traits
trait YamlExt {
    fn get_str(&self, key: &str) -> Result<&str, NVMSError>;
    fn get_opt<T>(&self, key: &str) -> Result<Option<T>, NVMSError>;
}

impl YamlExt for HashMap<Yaml, Yaml> {
    fn get_str(&self, key: &str) -> Result<&str, NVMSError> {
        self.get(&Yaml::from_string(key.to_string()))
            .and_then(|v| v.as_str())
            .ok_or_else(|| NVMSError::MissingField(key.to_string()))
    }

    fn get_opt<T>(&self, key: &str) -> Result<Option<T>, NVMSError> {
        Ok(self.get(&Yaml::from_string(key.to_string())).map(|v| v.clone()))
    }
}

impl NVMSConfig {
    pub fn from_str(content: &str) -> Result<Self, NVMSError> {
        let docs = yaml2::parse(content).map_err(|e| NVMSError::YamlError(e.to_string()))?;
        let doc = docs.first().ok_or(NVMSError::YamlError(
            "Empty document".to_string(),
        ))?;
        
        Self::try_from(doc.clone())
    }
}