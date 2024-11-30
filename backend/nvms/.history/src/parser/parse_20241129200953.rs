/**
 * Parse YAML Into OBJ
 * And Return The Object
 */
use config::Config;
use std::collections::HashMap;

// use src/lib/nvms.rs
use std::fs::File;
use crate::util::nvms::NVMS;



pub fn parse_config(nvmsFile: &std::fs::File) -> NVMSResult<NVMS> {
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
    let mut parsedConfig: NVMS = NVMS::default();
    let parseFile = Config::builder()
    .add_source(config::File::from(nvmsFile))
    .build()
    .unwrap();
    println!("{:?}", parseFile.try_deserialize::<HashMap<String,String>>());
    Ok(parsedConfig)
}


pub fn parse_and_validate_nvms(yaml: &str) -> Result<NVMSConfig, NVMSError> {
    let parsed = parse_config(yaml)?;
    validate_nvms(parsed)
}
