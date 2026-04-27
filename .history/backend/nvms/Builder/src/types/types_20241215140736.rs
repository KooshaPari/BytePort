use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Serialize, Deserialize)]
pub struct User {
    #[serde(default = "Uuid::new_v4")]
    pub uuid: Uuid,
    pub name: String,
    pub email: String,
    pub password: String, // Note: Consider using a password hash type
    pub aws_creds: AwsCreds,
    pub openai_creds: OpenAiCreds,
    pub portfolio: Portfolio,
    pub git: Git,
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub projects: Vec<Project>, // Assuming Project is defined elsewhere
    #[serde(skip_serializing_if = "Vec::is_empty", default)]
    pub instances: Vec<Instance>, // Assuming Instance is defined elsewhere
}

#[derive(Debug, Serialize, Deserialize)]
pub struct AwsCreds {
    pub access_key_id: String,
    pub secret_access_key: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct OpenAiCreds {
    pub api_key: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Portfolio {
    pub root_endpoint: String,
    pub api_key: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Git {
    pub access_token: String,
    pub refresh_token: String,
    pub token_expiry: DateTime<Utc>,
    pub refresh_token_expiry: DateTime<Utc>,
    #[serde(default)]
    pub repositories: Vec<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct LoginRequest {
    pub email: String,
    pub password: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct SignupRequest {
    pub name: String,
    pub email: String,
    pub password: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct LinkRequest {
    pub aws_creds: AwsCreds,
    pub openai_creds: OpenAiCreds,
    pub portfolio: Portfolio,
    pub git: Git,
}

// NVMS and related types
#[derive(Debug, Serialize, Deserialize)]
pub struct Nvms {
    pub name: String,
    pub description: String,
    pub services: Vec<Service>,
    pub aws: AwsConfig,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Service {
    pub name: String,
    pub path: String,
    pub build: Vec<String>,
    pub port: i32,
    pub env: Vec<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct AwsConfig {
    pub region: String,
    pub services: Vec<AwsServiceConfig>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct AwsServiceConfig {
    pub type_: String, // Using type_ as type is a reserved keyword in Rust
    pub engine: String,
    pub mode: String,
    pub replicas: i32,
    pub size: String,
    pub name: String,
    pub partitions: i32,
}

// Optional: Implementation blocks for the structs
impl User {
    pub fn new(name: String, email: String, password: String) -> Self {
        Self {
            uuid: Uuid::new_v4(),
            name,
            email,
            password,
            aws_creds: AwsCreds::default(),
            openai_creds: OpenAiCreds::default(),
            portfolio: Portfolio::default(),
            git: Git::default(),
            projects: Vec::new(),
            instances: Vec::new(),
        }
    }
}

// Implement Default for required structs
impl Default for AwsCreds {
    fn default() -> Self {
        Self {
            access_key_id: String::new(),
            secret_access_key: String::new(),
        }
    }
}

impl Default for OpenAiCreds {
    fn default() -> Self {
        Self {
            api_key: String::new(),
        }
    }
}

impl Default for Portfolio {
    fn default() -> Self {
        Self {
            root_endpoint: String::new(),
            api_key: String::new(),
        }
    }
}

impl Default for Git {
    fn default() -> Self {
        Self {
            access_token: String::new(),
            refresh_token: String::new(),
            token_expiry: Utc::now(),
            refresh_token_expiry: Utc::now(),
            repositories: Vec::new(),
        }
    }
}