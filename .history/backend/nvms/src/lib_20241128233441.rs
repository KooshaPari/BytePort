use spin_sdk::http::{IntoResponse, Request, Response};
use spin_sdk::http_component;




/// Deploy Component
#[http_component]
fn handle_deploy(req: Request) -> anyhow::Result<impl IntoResponse> {
    println!("Handling request to {:?}", req.header("spin-full-url"));
    Ok(Response::builder()
        .status(200)
        .header("content-type", "text/plain")
        .body("Hello, Fermyon")
        .build())
}
