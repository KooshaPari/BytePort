
use loco_rs::schema::table_auto_tz;
use sea_orm_migration::{prelude::*, schema::*};

#[derive(DeriveMigrationName)]
pub struct Migration;

#[async_trait::async_trait]
impl MigrationTrait for Migration {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        let table = table_auto_tz(UserMFA::Table)
            .col(
                ColumnDef::new(UserMFA::Id)
                    .string()
                    .primary_key()
                    .not_null()
                    .default("LOWER(HEX(RANDOMBLOB(4)) || '-' || HEX(RANDOMBLOB(2)) || '-' || '4' || SUBSTR(HEX(RANDOMBLOB(2)), 2) || '-' || SUBSTR('89AB', (ABS(RANDOM()) % 4) + 1, 1) || SUBSTR(HEX(RANDOMBLOB(2)), 2) || '-' || HEX(RANDOMBLOB(6)))")
            )
            .col(ColumnDef::new(UserMFA::UserID).string().not_null())
            .col(ColumnDef::new(UserMFA::MFAType).string().not_null())
            .col(ColumnDef::new(UserMFA::Secret).string().not_null())
            .col(ColumnDef::new(UserMFA::Enabled).integer().default(1).not_null())
            .col(
                ColumnDef::new(UserMFA::CreatedAt)
                    .timestamp_with_time_zone()
                    .default("CURRENT_TIMESTAMP")
                    .not_null(),
            )
            .col(
                ColumnDef::new(UserMFA::UpdatedAt)
                    .timestamp_with_time_zone()
                    .default("CURRENT_TIMESTAMP")
                    .not_null(),
            )
            .foreign_key(
                ForeignKey::create()
                    .from(UserMFA::Table, UserMFA::UserID)
                    .to(Users::Table, Users::Pid), // Foreign key referencing Users::Pid
            )
            .to_owned();

        manager.create_table(table).await?;
        Ok(())
    }

    async fn down(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .drop_table(Table::drop().table(UserMFA::Table).to_owned())
            .await
    }
}

#[derive(Iden)]
pub enum Users {
    Table,
    Id,
    Pid, // Assuming Pid is the UUID or unique identifier
    Email,
    FullName,
    Password,
    ApiKey,
    ResetToken,
    ResetSentAt,
    CreatedAt,
    UpdatedAt,
    LastLoginAt,
    FailedLoginAttempts
}

#[derive(Iden)]
pub enum UserMFA {
    Table,
    Id,
    UserID, // This will reference Users::Pid
    MFAType,
    Secret,
    Enabled,
    CreatedAt,
    UpdatedAt,
}
