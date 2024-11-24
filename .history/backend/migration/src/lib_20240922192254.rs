#![allow(elided_lifetimes_in_paths)]
#![allow(clippy::wildcard_imports)]

use sea_orm_migration::prelude::*;

mod m20230922_create_users_table;
mod m20230922_create_user_api_keys_table;
mod m20230922_create_user_mfa_table;

pub struct Migrator;

#[async_trait::async_trait]
impl MigrationTrait for Migrator {
    fn migrations() -> Vec<Box<dyn MigrationTrait>> {
        vec![
            Box::new(m20230922_create_users_table::Migration),
            Box::new(m20230922_create_user_api_keys_table::Migration),
            Box::new(m20230922_create_user_mfa_table::Migration),
        ]
    }
}
#![allow(elided_lifetimes_in_paths)]
#![allow(clippy::wildcard_imports)]
pub use sea_orm_migration::prelude::*;
mod m20230922_create_users_table;
mod m20230922_create_user_api_keys_table;
mod m20230922_create_user_mfa_table;

pub struct Migrator;

#[async_trait::async_trait]
impl MigratorTrait for Migrator {
    fn migrations() -> Vec<Box<dyn MigrationTrait>> {
        vec![
            Box::new(m20230922_create_users_table::Migration),
            Box::new(m20230922_create_user_api_keys_table::Migration),
            Box::new(m20230922_create_user_mfa_table::Migration),
        ]
    }
}