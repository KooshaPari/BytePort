
mod lib;
fn main() {
   let nvms_file: String = include_str!("/examples/byteport.nvms").to_string();
   let nvms = lib::parse_config(nvms_file).unwrap();
}