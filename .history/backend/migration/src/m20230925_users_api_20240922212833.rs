
use sea_orm_migration::prelude::*;
use crate::users_enum::Users; 
#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .create_table(
                Table::create()
                    .table(UserAPIKeys::Table)
                    .if_not_exists()
                    .col(
                        ColumnDef::new(UserAPIKeys::Id)
                            .string()
                            .primary_key()
                            .not_null()
                            .default("LOWER(HEX(RANDOMBLOB(4)) || '-' || HEX(RANDOMBLOB(2)) || '-' || '4' || SUBSTR(HEX(RANDOMBLOB(2)), 2) || '-' || SUBSTR('89AB', (ABS(RANDOM()) % 4) + 1, 1) || SUBSTR(HEX(RANDOMBLOB(2)), 2) || '-' || HEX(RANDOMBLOB(6)))")
                    )
                    .col(ColumnDef::new(UserAPIKeys::UserID).string().not_null())
                    .col(ColumnDef::new(UserAPIKeys::APIProvider).string().not_null())
                    .col(ColumnDef::new(UserAPIKeys::AccessKeyID).string().not_null())
                    .col(ColumnDef::new(UserAPIKeys::SecretAccessKey).string().not_null())
                    .col(ColumnDef::new(UserAPIKeys::CreatedAt).timestamp_with_time_zone().default("CURRENT_TIMESTAMP").not_null())
                    .col(ColumnDef::new(UserAPIKeys::UpdatedAt).timestamp_with_time_zone().default("CURRENT_TIMESTAMP").not_null())
                    .foreign_key(
                        ForeignKey::create()
                            .from(UserAPIKeys::Table, UserAPIKeys::UserID)
                            .to(Users::Table, Users::Pid),  // Foreign key referencing Users::Pid
                    )
                    .to_owned(),
            )
            .await
    }

    async fn down(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager.drop_table(Table::drop().table(UserAPIKeys::Table).to_owned()).await
    }
}

#[derive(Iden)]
pub enum UserAPIKeys {
    Table,
    Id,
    UserID,
    APIProvider,
    AccessKeyID,
    SecretAccessKey,
    CreatedAt,
    APOUpdatedAt,
}
