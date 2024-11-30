
mod util;
mod parser;
use parser::parse;
use std::fs::File;
use spin_sdk::http::{IntoResponse, Request, Response};
use spin_sdk::http_component;

#[http_component]
fn handle_test_run(req: Request) -> anyhow::Result<impl IntoResponse> {
   println!("Hello, world!");
   let nvms_file: std::fs::File =  File::open("examples/byteport.yaml").unwrap();
   let nvms = parse::parse_config(&nvms_file);
   println!("Done Parsing");
   Ok(Response::builder()
        .status(200)
        .header("content-type", "text/plain")
        .body("Hello, Fermyon")
        .build())
}

