use spin_sdk::http::{IntoResponse, Request, Response};
use spin_sdk::http_component;


#[http_component]
fn handle_builder(req: Request) -> anyhow::Result<impl IntoResponse> {
    println!("Handling request to {:?}", req.header("spin-full-url"));
    // We'll Part out the original MVP Deploy Process as Such:
    /** Gin->SpinMain (The routine will be run at a high level from here)
     *  -> Provisioner (Find codebase and needed files, post each service to an s3 bucket, return)
)     *  -> Builder ( Po)
     */

    Ok(Response::builder()
        .status(200)
        .header("content-type", "text/plain")
        .body("Hello, Fermyon")
        .build())
}
