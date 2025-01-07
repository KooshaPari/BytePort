package ec2

import (
	"context"
	"encoding/xml"
	"fmt"
)

/*
https://iam.amazonaws.com/?Action=CreateInstanceProfile
&InstanceProfileName=Webserver
&Path=/application_abc/component_xyz/
&Version=2010-05-08
&AUTHPARAMS*/
func(c *Client) GetInstanceProfile(ctx context.Context, name string)(*GetInstanceProfileResponse, error){
	fmt.Println("Getting instance profile: '", name+"'")
	params := map[string]string{
		"Action": "GetInstanceProfile",
		"InstanceProfileName": name,
		"Version": "2010-05-08",
	}
	req, err := c.newRequest(ctx, "GET", params, nil )
	if err != nil {
		return nil, err
	}
	fmt.Println("Request: ", req)
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	var result GetInstanceProfileResponse
	if err := xml.NewDecoder(body).Decode(result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}
func(c *Client) CreateInstanceProfile(ctx context.Context, name string) (*CreateInstanceProfileResponse, error) {
	fmt.Println("Creating instance profile")
	params := map[string]string{
		"Action": "CreateInstanceProfile",
		"InstanceProfileName": name, 
		"Version": "2010-05-08",
	}
	req, err := c.newRequest(ctx, "GET", params, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result CreateInstanceProfileResponse
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}
func (c *Client) DeleteInstanceProfile(ctx context.Context, name string) (*DeleteInstanceProfileResponse, error) {
	params := map[string]string{
		"Action": "DeleteInstanceProfile",
		"InstanceProfileName": name,
	}
	req, err := c.newRequest(ctx, "GET", params, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result DeleteInstanceProfileResponse
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}