
impl MigrationTrait for Migration {
    async fn up(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager
            .create_table(
                Table::create()
                    .table(Users::Table)
                    .if_not_exists()
                    .col(
                        ColumnDef::new(Users::Id)
                            .string()
                            .primary_key()
                            .not_null()
                            .default("LOWER(HEX(RANDOMBLOB(4)) || '-' || HEX(RANDOMBLOB(2)) || '-' || '4' || SUBSTR(HEX(RANDOMBLOB(2)), 2) || '-' || SUBSTR('89AB', (ABS(RANDOM()) % 4) + 1, 1) || SUBSTR(HEX(RANDOMBLOB(2)), 2) || '-' || HEX(RANDOMBLOB(6)))")
                    )
                    .col(ColumnDef::new(Users::FullName).string().null())
                    .col(ColumnDef::new(Users::Email).string().not_null().unique_key())
                    .col(ColumnDef::new(Users::PasswordHash).string().not_null())
                    .col(ColumnDef::new(Users::PasswordSalt).string().null())
                    .col(ColumnDef::new(Users::PasswordResetToken).string().null())
                    .col(ColumnDef::new(Users::PasswordResetTokenExpiry).string().null())
                    .col(ColumnDef::new(Users::IsActive).integer().not_null().default(1))
                    .col(ColumnDef::new(Users::CreatedAt).timestamp_with_time_zone().default("CURRENT_TIMESTAMP").not_null())
                    .col(ColumnDef::new(Users::UpdatedAt).timestamp_with_time_zone().default("CURRENT_TIMESTAMP").not_null())
                    .col(ColumnDef::new(Users::LastLoginAt).timestamp_with_time_zone().null())
                    .col(ColumnDef::new(Users::FailedLoginAttempts).integer().default(0))
                    .to_owned(),
            )
            .await
    }

    async fn down(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager.drop_table(Table::drop().table(Users::Table).to_owned()).await
    }
}

#[derive(Iden)]
pub enum Users {
    Table,
    Id,
    FullName,
    Email,
    PasswordHash,
    PasswordSalt,
    PasswordResetToken,
    PasswordResetTokenExpiry,
    IsActive,
    CreatedAt,
    UpdatedAt,
    LastLoginAt,
    FailedLoginAttempts,
}
