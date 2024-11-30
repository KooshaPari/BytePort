
mod util;
mod parser;
use parser::parse;
use std::fs::File;
fn main() {
   println!("Hello, world!");
   let nvms_file: std::fs::File =  File::open("examples/byteport.yaml").unwrap();
   let nvms = parse::parse_config(&nvms_file);
   println!("Done Parsing");
}

main()