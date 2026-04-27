use sea_orm_migration::prelude::*;

#[tokio::main]
async fn main() -> Result<(), DbErr> {
    let db = sea_orm::Database::connect("sqlite://your_database.sqlite").await?;
}