use spin_sdk::http::{IntoResponse, Request, Response};
use spin_sdk::http_component;




/// Deploy Component
#[http_component]
fn handle_deploy(req: Request) -> anyhow::Result<impl IntoResponse> {
   /**
    *  1. get user and proj object
    * let user = req.b
    * let config = locateNVMS()
    let projectNVMS: NVMS = parseAndValidateNVMS(config)
    let project = Project::new(projectNVMS)
    let instance = await buildAnddeploy(project)
    let templates = await getTemplates()    
    

    */
   
   
   todo!()
}