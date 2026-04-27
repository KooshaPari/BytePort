use sea_orm_migration::{prelude::*, schema::*};
use crate::users_enum::Users; 
#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {

    }

    async fn down(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
                    .drop_table(Table::drop().table(Users::Table).to_owned())
                    .await;
        Ok(());
    }
}

