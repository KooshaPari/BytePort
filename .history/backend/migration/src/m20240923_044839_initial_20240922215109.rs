use sea_orm_migration::{prelude::*, schema::*};
use crate::users_enum::Users; 
#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
       
        /*
                let table = table_auto_tz(Users::Table)
                    
                    .col(uuid(Users::Pid))
                    .col(ColumnDef::new(Users::FullName).string().null())
                    .col(string_uniq(Users::Email))
                    .col(string(Users::Password))
                    .col(string(Users::ApiKey).unique_key())
                    .col(string_null(Users::ResetToken))
                    .col(timestamp_with_time_zone_null(Users::ResetSentAt))
                    
                    .col(ColumnDef::new(Users::CreatedAt).timestamp_with_time_zone().default("CURRENT_TIMESTAMP").not_null())
                            .col(ColumnDef::new(Users::UpdatedAt).timestamp_with_time_zone().default("CURRENT_TIMESTAMP").not_null())
                            .col(ColumnDef::new(Users::LastLoginAt).timestamp_with_time_zone().null())
                            .col(ColumnDef::new(Users::FailedLoginAttempts).integer().default(0))
                    .to_owned();
                manager.create_table(table).await?;
                Ok(())
            }

  


        */
        todo!()
    }

    async fn down(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
                    .drop_table(Table::drop().table(Users::Table).to_owned())
                    .await
        Ok(())
    }
}

