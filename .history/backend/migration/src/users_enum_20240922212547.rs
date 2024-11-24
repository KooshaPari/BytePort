use sea_orm_migration::prelude::*;
#[derive(Iden)]
pub enum Users {
    Table,
    Id,
    Pid,
    Email,
    FullName,
    Password,
    ApiKey,
    ResetToken,
    ResetSentAt,
    CreatedAt,
    UpdatedAt,
    LastLoginAt,
    FailedLoginAttempts,
}
