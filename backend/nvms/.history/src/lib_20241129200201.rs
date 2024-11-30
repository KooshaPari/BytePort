
mod util;
fn main() {
   let nvms_file: String = include_str!("./examples/byteport.nvms").to_string();
   let nvms = util::nvms::parse_config(nvms_file).unwrap();
}