/**
 * Parse YAML Into OBJ
 * And Return The Object
 */

use yaml_rust2::{YamlLoader, YamlEmitter};
use std::collections::HashMap;
use thiserror::Error;
use std::str::FromStr;
// use src/lib/nvms.rs
use crate::nvms::NVMS;

/*
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
} */

pub fn parse_config(yaml: &str) -> NVMS, NVMSError> {
    /*
    Grab Header (FROM,NAME,DESCR,VERSION,PROJECT)
    Read Templates (Template (Type) (Presets))
    Read Clusters (Cluster (Type) (PRESET | RESOURCES) CONFIG(INSTANCES PATH BUILD SCALE HEALTH ENV )))
    Read Services(Service (PATH BUILD PORT ENV PROTOCOLS (PRESET | RESOURCES))) 
    Read AWS Config (Region, MultiRegion?, VPC, Services)
    NETWORK ( DOMAIN SSL LOADBALANCER CDN SECURITY)
    MONITORING(Provider, Metrics, Alerts, Logging, Tracing) 
    DEPLOYMENT (Strategy, Batch Size, Health_Check_Grace, Tiemout, Rollback)
     BACKUP (Enabled, Retention, Schedule, Destinations, )
     MAINTENANCE (Updates(security, system, schedule) Patching)

     

     */
    let configFile = &(YamlLoader::load_from_str(s).unwrap())[0];
    let from = configFile["from"].as_str().unwrap();
    let name = configFile["name"].as_str().unwrap();
    let description = configFile["description"].as_str().unwrap();
    let version = configFile["version"].as_str().unwrap();
    let project = configFile["project"].as_str().unwrap();
    let header = HeaderConfig {
        from: from.to_string(),
        name: name.to_string(),
        description: description.to_string(),
        version: version.to_string(),
        project: project.to_string(),
    };
    // PARSE HEADER

    
    

    todo!()

}/*
pub fn validate_nvms(config: HashMap<String, Yaml>) -> Result<NVMSConfig, NVMSError> {
    todo!()
}

pub fn parse_and_validate_nvms(yaml: &str) -> Result<NVMSConfig, NVMSError> {
    let parsed = parse_config(yaml)?;
    validate_nvms(parsed)
} */