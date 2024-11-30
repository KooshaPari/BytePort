
mod util;
mod parser;
use parser::parse;
use std::fs::File;
fn main() {
   let nvms_file: std::fs::File =   
   let nvms = parse::parse_config(nvms_file).unwrap();
}