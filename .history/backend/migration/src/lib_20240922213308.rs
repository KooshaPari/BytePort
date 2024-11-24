#![allow(elided_lifetimes_in_paths)]
#![allow(clippy::wildcard_imports)]
pub use sea_orm_migration::prelude::*;
pub mod users_enum;
pub mod migrator;
mod m20230925_users;
//mod m20230925_users_mfa;
//mod m20230925_users_api;pub struct Migrator;

#[async_trait::async_trait]
impl MigratorTrait for Migrator {
    fn migrations() -> Vec<Box<dyn MigrationTrait>> {
        vec![
            Box::new(m20230925_users::Migration), 
            //Box::new(m20230925_users_mfa::Migration), 
            Box::new(m20230925_users_api::Migration), 
        ]
    }
}
