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

pub fn parse_yaml(yaml: &str) -> Result<HashMap<String, Yaml>, NVMSError> {

    todo!()

}
pub fn validate_nvms(config: HashMap<String, Yaml>) -> Result<NVMSConfig, NVMSError> {
    todo!()
}

