pub struct S;

impl volo_gen::bytebridge::NanoVms for S {
    async fn start_vm(
        &self,
        _req: ::volo_grpc::Request<volo_gen::bytebridge::StartVmRequest>,
    ) -> ::std::result::Result<
        ::volo_grpc::Response<volo_gen::bytebridge::VmResponse>,
        ::volo_grpc::Status,
    > {
        ::std::result::Result::Ok(::volo_grpc::Response::new(Default::default()))
    }

    async fn stop_vm(
        &self,
        _req: ::volo_grpc::Request<volo_gen::bytebridge::StopVmRequest>,
    ) -> ::std::result::Result<
        ::volo_grpc::Response<volo_gen::bytebridge::VmResponse>,
        ::volo_grpc::Status,
    > {
        ::std::result::Result::Ok(::volo_grpc::Response::new(Default::default()))
    }
}
