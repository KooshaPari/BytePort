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

use crate::{
    controllers,  // Assuming controllers other than notes and auth are present
    tasks,
    workers::downloader::DownloadWorker,
};

pub struct App;

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
            // Remove routes related to auth and notes
            // Add any other routes you have, if applicable
            //.add_route(controllers::some_other_route())
    }

    fn connect_workers<'a>(p: &'a mut Processor, ctx: &'a AppContext) {
        p.register(DownloadWorker::build(ctx));
    }

    fn register_tasks(tasks: &mut Tasks) {
        tasks.register(tasks::seed::SeedData);
    }

    // Removing the truncate logic related to users and notes
    async fn truncate(_db: &DatabaseConnection) -> Result<()> {
        // Implement your own table truncation logic here if needed
        Ok(())
    }

    // Removing seed logic related to users and notes
    async fn seed(_db: &DatabaseConnection, _base: &Path) -> Result<()> {
        // Implement your own seed logic here if needed
        Ok(())
    }
}