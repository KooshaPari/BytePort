
        manager
            .create_table(
                Table::create()
                    .table(UserMFA::Table)
                    .if_not_exists()
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
                    .col(ColumnDef::new(UserMFA::CreatedAt).timestamp_with_time_zone().default("CURRENT_TIMESTAMP").not_null())
                    .col(ColumnDef::new(UserMFA::UpdatedAt).timestamp_with_time_zone().default("CURRENT_TIMESTAMP").not_null())
                    .foreign_key(
                        ForeignKey::create()
                            .from(UserMFA::Table, UserMFA::UserID)
                            .to(Users::Table, Users::Id),
                    )
                    .to_owned(),
            )
            .await
    }

    async fn down(&self, manager: &SchemaManager) -> Result<(), DbErr> {
        manager.drop_table(Table::drop().table(UserMFA::Table).to_owned()).await
    }
}

#[derive(Iden)]
pub enum UserMFA {
    Table,
    Id,
    UserID,
    MFAType,
    Secret,
    Enabled,
    CreatedAt,
    UpdatedAt,
}
