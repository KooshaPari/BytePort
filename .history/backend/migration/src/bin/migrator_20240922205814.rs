use sea_orm_migration::prelude::*;

#[tokio::main]
async fn main() {
    let db = Database::connect("").await.expect("Failed to connect to database");
    Migrator::up(&db, None).await.expect("Migration failed");
}