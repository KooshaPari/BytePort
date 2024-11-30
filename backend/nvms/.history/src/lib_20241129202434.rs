
mod util;
mod parser;
use parser::parse;
use std::fs::File;
use spin_sdk::http::IntoResponse;
use spin_sdk::http_component;

#[http_component]
fn main() {
   println!("Hello, world!");
   let nvms_file: std::fs::File =  File::open("examples/byteport.yaml").unwrap();
   let nvms = parse::parse_config(&nvms_file);
   println!("Done Parsing");
}

