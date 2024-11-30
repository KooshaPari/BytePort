use spin_sdk::http::{IntoResponse, Request, Response};
use spin_sdk::http_component;
mod ;




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
    project.instances.push(instance)
    let portDets = Portfolio::new({Decrypt(user.Portfolio.rootEndpoint), Decrypt(user.Portfolio.apiKey)})
    let templates = await getTemplates(portDets)
    let processed = buildFromTemplates(templates, project)
    let response = await sendProcessed(processed)
    if(response is ok) {return http.StatusOK} else {return Http.statusErr, err}    
    

    */
   
   
   todo!()
}
