

mod gen {
    include!(concat!(env!("OUT_DIR"), "/volo_gen.rs"));
}

pub use gen::volo_gen::*;
