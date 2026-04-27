/**
 * Parse YAML Into OBJ
 * And Return The Object
 */

use yaml2::Yaml;
use std::collections::HashMap;
use thiserror::Error;
use std::str::FromStr;

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

pub fn parse_config(yaml: &str) -> Result<HashMap<String, Yaml>, NVMSError> {
    /*
    Grab Header (FROM,NAME,DESCR,VERSION,PROJECT)
    Read Templates (Template (Type) (Presets))
    Read Clusters (Cluster (Type) (PRESET | RESOURCES) CONFIG(INSTANCES PATH BUILD SCALE HEALTH ENV )))
        
     */
    todo!()

}
pub fn validate_nvms(config: HashMap<String, Yaml>) -> Result<NVMSConfig, NVMSError> {
    todo!()
}

pub fn parse_and_validate_nvms(yaml: &str) -> Result<NVMSConfig, NVMSError> {
    let parsed = parse_config(yaml)?;
    validate_nvms(parsed)
}
