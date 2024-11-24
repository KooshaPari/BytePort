use std::path::Path;

use async_trait::async_trait;
use loco_rs::{
    app::{AppContext, Hooks},
    boot::{create_app, BootResult, StartMode},
    controller::AppRoutes,
    db::{self, truncate_table},
    environment::Environment,
    task::Tasks,
    worker::{AppWorker, Processor},
    Result,
};
use migration::Migrator;
use sea_orm::DatabaseConnection;

use crate::{
    controllers,
    //models::_entities::{notes, users},
    //tasks,
    //workers::downloader::DownloadWorker,
};

pub struct App;
pub async fn connect() -> DatabaseConnection {
    let db_url = "sqlite://./backend_development.db";
    let db = Database::connect(db_url).await.unwrap();
    db
}

#[async_trait]
impl Hooks for App {
    fn app_name() -> &'static str {
        env!("CARGO_CRATE_NAME")
    }

    fn app_version() -> String {
        format!(
            "{} ({})",
            env!("CARGO_PKG_VERSION"),
            option_env!("BUILD_SHA")
                .or(option_env!("GITHUB_SHA"))
                .unwrap_or("dev")
        )
    }

    async fn boot(mode: StartMode, environment: &Environment) -> Result<BootResult> {
        create_app::<Self, Migrator>(mode, environment).await
    }

    fn routes(_ctx: &AppContext) -> AppRoutes {
        AppRoutes::with_default_routes()
            .prefix("/api")

    }

    fn connect_workers<'a>(p: &'a mut Processor, ctx: &'a AppContext) {
       // p.register(DownloadWorker::build(ctx));
       todo!()
    }

    fn register_tasks(tasks: &mut Tasks) {
        todo!()
    }

    async fn truncate(db: &DatabaseConnection) -> Result<()> {

        Ok(())
    }

    async fn seed(db: &DatabaseConnection, base: &Path) -> Result<()> {
        
        Ok(())
    }
}