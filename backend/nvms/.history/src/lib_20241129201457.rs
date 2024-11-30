
mod util;
mod parser;
use parser::parse;
use std::
fn main() {
   let nvms_file: std::fs::File = include_str!("./examples/byteport.nvms").to_string();
   let nvms = parse::parse_config(nvms_file).unwrap();
}