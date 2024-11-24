

        manager
            .create_table(
                Table::create()
                    .table(Users::Table)
                    .if_not_exists()
                    string().null())
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
