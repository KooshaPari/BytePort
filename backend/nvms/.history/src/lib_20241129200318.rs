
mod util;
mod parser;
use parser::parse;

fn main() {
   let nvms_file: String = include_str!("./examples/byteport.nvms").to_string();
   let nvms = parse::parse_config(nvms_file).unwrap();
}