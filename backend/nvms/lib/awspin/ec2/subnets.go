package ec2

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
)


func(c *Client) DescribeSubnets(ctx context.Context, vpcId string) (*DescribeSubnetsResponse , error) {
 
    params := map[string]string{
        "Action": "DescribeSubnets",
        "Filter.1.Name": "vpc-id",
        "Filter.1.Value.1": vpcId,
 
        "Version": "2016-11-15",
    }
    req, err := c.newRequest(ctx, "GET", params, nil)
    if err != nil {
        return nil, err
    }
    resp, err := c.do(req)
    if err != nil {
        fmt.Println("Error getting Subnet: ", err)
        return nil, err
    }
    defer resp.Body.Close()
    bodyBytes, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %w", err)
    }
    //fmt.Println("Response Body:", string(bodyBytes))

    // Reset the response body for decoding
    resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
    var subnets DescribeSubnetsResponse
    if err := xml.NewDecoder(resp.Body).Decode(&subnets); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    return &subnets, nil
}
