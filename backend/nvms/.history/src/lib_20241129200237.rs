
mod util;
m
fn main() {
   let nvms_file: String = include_str!("./examples/byteport.nvms").to_string();
   let nvms = parser::parse_config(nvms_file).unwrap();
}