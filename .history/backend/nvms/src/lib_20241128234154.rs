use spin_sdk::http::{IntoResponse, Request, Response};
use spin_sdk::http_component;




/// Deploy Component
#[http_component]
fn handle_deploy(req: Request) -> anyhow::Result<impl IntoResponse> {
   /**
    *  1. get user and proj object
    * let user = req.body().user
    * let proj = req.body().proj
    * let config = locateNVMS()
    let projectNVMS: NVMS = parseAndValidateNVMS(config)
    let project = Project::new(projectNVMS)
    let instance = await buildAnddeploy(project)
    let portDets = Portfolio::new({Decrypt(user.Portfolio.rootEndpoint),})
    let templates = await getTemplates(user.Portfolio)    
    

    */
   
   
   todo!()
}
