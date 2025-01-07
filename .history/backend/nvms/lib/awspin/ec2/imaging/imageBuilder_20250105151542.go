package imaging

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"nvms/models"
	"time"
)
 
func (c *Client) CreatePipeline(ctx context.Context, name string, infraArn string, recipeArn string) (*CreatePipelineResponse, error) {
 
	reqbody := CreatePipelineRequest{
		ClientToken: time.Now().String(),
		InfrastructureConfigurationArn: infraArn,
		Name: name,
		ContainerRecipeArn: recipeArn,
	}
 
	 body, err := json.Marshal(reqbody); 
	 if err != nil {
		return nil, err
	}
 
	req, err := c.newRequest(ctx, "PUT", nil, body,"/CreateImagePipeline")
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	 
	rbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	var result CreatePipelineResponse
	if err := json.Unmarshal(rbody, result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
	
}
func (c *Client) ExecuteImgPipeline(ctx context.Context, pipelineArn string) (*ExecuteImgPipelineResponse, error) {
	reqBody := ExecuteImgPipelineRequest{
		ClientToken: time.Now().String(),
		ImagePipelineArn: pipelineArn,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	params := map[string]string{
		"Version": "2019-12-02",
		"Action": "StartExecuteImagePipeline",}
 
	req, err := c.newRequest(ctx, "POST", params, body,"/StartImagePipelineExecution")
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
 
	rbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	var result ExecuteImgPipelineResponse
	if err := json.Unmarshal(rbody, result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}
 
func (c *Client) CreateInfrastructureConfiguration(ctx context.Context, name string, instProfName string) (*CreateInfrastructureConfigurationResponse, error){
	// append /CreateInfrastructureConfiguration to the URL
	fmt.Println("Creating Infrastructure Configuration")
 
	reqBody := CreateInfrastructureConfigurationRequest{
		ClientToken: time.Now().String(),
		Description: "Infrastructure Configuration for Image Builder",
		InstanceProfileName: instProfName,
		Name: name,}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	params := map[string]string{
		"Version": "2019-12-02",
		"Action": "CreateInfrastructureConfiguration",}
	req, err := c.newRequest(ctx, "PUT", params,body,"/CreateInfrastructureConfiguration")
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	// revert the URL back to the original
 
	defer resp.Body.Close()
	rbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	var result CreateInfrastructureConfigurationResponse
	if err := json.Unmarshal(rbody,result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}
 
func (c *Client) DeleteInfrastructureConfiguration(ctx context.Context, name string, infraARN string) ( error){
 
	params := map[string]string{
		"clientToken": time.Now().String(),
		"infrastructureConfigArn": infraARN,
		"Version": "2019-12-02",
"Action": "DeleteInfrastructureConfiguration",
 
	}
	req, err := c.newRequest(ctx, "DELETE", params, nil, "/DeleteInfrastructureConfiguration")
	if err != nil {
		return  err
	}

	resp, err := c.do(req)
	if err != nil {
		return  err
	}
 
	defer resp.Body.Close()
	rbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return   fmt.Errorf("failed to read response body: %w", err)
	}
	var result DeleteInfrastructureConfigurationResponse
	if err := json.Unmarshal(rbody, result); err != nil {
		return   fmt.Errorf("failed to decode response: %w", err)
	}
	return  nil
}
 
func (c *Client) CreateImageRecipe(ctx context.Context, components []models.ImageComponent, name string) (*CreateImageRecipeResponse, error){
	request := CreateImageRecipeRequest{
		Components: components,
		ClientToken: time.Now().String(),
		SemanticVersion: "1.0.0",
		Name: name,
		ParentImage: "ami-0d498bdc7202f4aa6",
 
}
		

		body,err := json.Marshal(request)
		if err != nil {
			return nil, err
		}
 
	params := map[string]string{
		"Version": "2019-12-02",
"Action": "CreateImageRecipe",}
	req, err := c.newRequest(ctx, "PUT", params,body, "/CreateImageRecipe")
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	c.endpointURL = c.endpointURL[:len(c.endpointURL)-18]
	rbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	var result CreateImageRecipeResponse
	if err := json.Unmarshal(rbody, result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil

}
//DELETE /DeleteImageRecipe?imageRecipeArn=imageRecipeArn HTTP/1.1
func (c *Client) DeleteImageRecipe(ctx context.Context, targetArn string ) ( error){
	c.endpointURL += "/DeleteImageRecipe"
	params := map[string]string{
		"imageRecipeArn": targetArn,
		"Version": "2019-12-02",
"Action": "DeleteImageRecipe",
	}
	req, err := c.newRequest(ctx, "DELETE", params,nil,)
	if err != nil {
		return   err
	}
	resp, err := c.do(req)
	if err != nil {
		return   err
	}
	defer resp.Body.Close()
	c.endpointURL = c.endpointURL[:len(c.endpointURL)-18]
	rbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return   fmt.Errorf("failed to read response body: %w", err)
	}
	var result CreateImageRecipeResponse
	if err := json.Unmarshal(rbody, result); err != nil {
		return   fmt.Errorf("failed to decode response: %w", err)
	}
	return  nil

}
func (c *Client) CreateImageComponent(ctx context.Context, name string, version string, platform string, data string) (*CreateImageComponentResponse, error){
	request := CreateImageComponentRequest{
		Name: name,
		ClientToken: time.Now().String(),
		Platform: platform,
		SemanticVersion: "1.0.0",
		Data: data,
		 
	}
	body,err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	c.endpointURL += "/CreateComponent"
	params := map[string]string{
		"Version": "2019-12-02",
"Action": "CreateComponent",}

	req, err := c.newRequest(ctx, "PUT", params,body)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	c.endpointURL = c.endpointURL[:len(c.endpointURL)-13]
	var result CreateImageComponentResponse
	if err := json.Unmarshal(rbody, result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil

}
func (c *Client) CreateContainerRecipe(ctx context.Context,components []ComponentsConfig, name string, targetRepo ContainerRepo, workDir string)(CreateContainerRecipeResponse, error){
	request := CreateContainerRecipeRequest{
		Components: components,
		ClientToken: time.Now().String(),
		SemanticVersion: "1.0.0",
		ContainerType: "DOCKER",
		TargetRepository: targetRepo,
		Name: name,
		ParentImage: "ami-0d498bdc7202f4aa6",
		WorkingDirectory: workDir,
	}
	body,err := json.Marshal(request)
	if err != nil {
		return CreateContainerRecipeResponse{}, err
	}
	c.endpointURL += "/CreateContainerRecipe"
	params := map[string]string{
		"Version": "2019-12-02",
"Action": "CreateContainerRecipe",
}
	req, err := c.newRequest(ctx, "PUT", params,body)
	if err != nil {
		return CreateContainerRecipeResponse{}, err
	}
	resp, err := c.do(req)
	if err != nil {
		return CreateContainerRecipeResponse{}, err
	}
	defer resp.Body.Close()
	rbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return CreateContainerRecipeResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}
	c.endpointURL = c.endpointURL[:len(c.endpointURL)-19]
	var result CreateContainerRecipeResponse
	if err := json.Unmarshal(rbody, result); err != nil {
		return CreateContainerRecipeResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}
	return result, nil
}