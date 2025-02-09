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
use config::Config;


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
    let nvmsConfig = Config::builder()
    .add_source(Config::default())
    .add_source(Config::File::with_("byteport"))
    .build().unwrap();
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
    let enum ConfigType {
        Header,
        Templates,
        Clusters,
        Services,
        AWS,
        Network,
        Monitoring,
        Deployment,
        Backup,
        Maintenance,
    }
    let enum headerFields {
        from,
        name,
        description,
        version,
        project,
    }


    for field in configFile {
        // increment thru fields
        // for each field go into a a sub-loop of sub fields, nested as far as needed


    }




    // PARSE HEADER
    let mut parsedHeader: HeaderConfig = HeaderConfig {
        from: "".to_string(),
        name: "".to_string(),
        description: "".to_string(),
        version: "".to_string(),
        project: "".to_string(),
    };
    
    

    todo!()

}/*
pub fn validate_nvms(config: HashMap<String, Yaml>) -> Result<NVMSConfig, NVMSError> {
    todo!()
}

pub fn parse_and_validate_nvms(yaml: &str) -> Result<NVMSConfig, NVMSError> {
    let parsed = parse_config(yaml)?;
    validate_nvms(parsed)
} */
