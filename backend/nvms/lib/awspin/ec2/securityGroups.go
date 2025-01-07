package ec2

import (
	"context"
	"encoding/xml"
	"fmt"
)
func(c *Client) DescribeSecurityGroups(ctx context.Context, vpcId string) (*DescribeSecurityGroupsResponse, error) { 
    /*
    curl "https://ec2.us-east-1.amazonaws.com/?Action=DescribeSecurityGroups&Filter.1.Name=vpc-id&Filter.1.Value.1=vpc-xxxxx&Version=2016-11-15" \
-H "Content-Type: application/x-www-form-urlencoded" \
--aws-sigv4 "aws:amz:us-east-1:ec2" \
--user "YOUR_ACCESS_KEY:YOUR_SECRET_KEY"*/
   params := map[string]string{
        "Action": "DescribeSecurityGroups",
        "Filter.1.Name": "vpc-id",
        "Filter.1.Value.1": vpcId,
        // filter by azn e.g. us-east 1
 
        "Version": "2016-11-15",}
    req, err := c.newRequest(ctx, "GET", params, nil)
    if err != nil {
        return nil, err
    }
    resp, err := c.do(req)
    if err != nil {
        fmt.Println("Error getting VPC: ", err)
        return nil, err
    }
    defer resp.Body.Close()
    
    var securityGroups DescribeSecurityGroupsResponse
    if err := xml.NewDecoder(resp.Body).Decode(&securityGroups); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    return &securityGroups, nil
}