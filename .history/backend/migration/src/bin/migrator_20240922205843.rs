use sea_orm_migration::prelude::*;

#[tokio::main]
 fn main() {
    let db = Database::connect("../backend_database.sqlite").await.expect("Failed to connect to database");
    Migrator::up(&db, None).await.expect("Migration failed");
}