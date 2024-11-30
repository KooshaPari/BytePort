use spin_sdk::http::{IntoResponse, Request, Response};
use spin_sdk::http_component;
mod lib;

fn main() {
   let nvms_file: String = include_str!("/examples/byteport.nvms").to_string();
   let nvmsRe = lib::parse_config(nvms_file).unwrap();
}