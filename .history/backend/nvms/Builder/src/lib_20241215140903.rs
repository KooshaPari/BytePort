use spin_sdk::http::{IntoResponse, Request, Response};
use spin_sdk::http_component;
use crate::types::*;

#[http_component]
fn handle_builder(req: Request) -> anyhow::Result<impl IntoResponse> {
    println!("Handling request to {:?}", req.header("spin-full-url"));
    // We'll Part out the original MVP Deploy Process as Such:
    /** Gin->SpinMain (The routine will be run at a high level from here)
     *  -> Provisioner (Find codebase and needed files, return)
)     *  -> Builder ( Post each service to S3 Bucket, Deploy EC2 and build services, return)
*     -> Done
        Builder Will Be Given:
        - User(Creds, Info), - NVMS Config, - ZipBall Codebase, - Project Name
        We Will Start Off with JUST S3 Ops, Refactoring all code up until that point below.
        
     */
    let mut user = User::default();
    let mut nvmsConfig = NVMSConfig::default();
    let mut zipball = Zipball::default();
    let mut projectName = ProjectName::default();

    Ok(Response::builder()
        .status(200)
        .header("content-type", "text/plain")
        .body("Hello, Fermyon")
        .build())
}
