package ec2

func(c *Client) describeDefaultVPC(ctx context.Context) (string, error) {
    /*
    curl "https://ec2.us-east-1.amazonaws.com/?Action=DescribeVpcs&Filter.1.Name=isDefault&Filter.1.Value.1=true&Version=2016-11-15" \
-H "Content-Type: application/x-www-form-urlencoded" \
--aws-sigv4 "aws:amz:us-east-1:ec2" \
--user "YOUR_ACCESS_KEY:YOUR_SECRET_KEY"
    */
    fmt.Println("getting vpc")
    params := map[string]string{
        "Action": "DescribeVpcs",
        "Filter.1.Name": "isDefault",
        "Filter.1.Value.1": "true", 
        "Version": "2016-11-15",}
    req, err := c.newRequest(ctx, "GET", params, nil)
    if err != nil {
        fmt.Println("Err Getting VPC Req: ", err)
        return "", err
    }
    resp, err := c.do(req)
    if err != nil {
        fmt.Println("Error getting VPC: ", err)
        return "", err
    }
 
    defer resp.Body.Close()
    var defaultVPC  DescribeVpcsResponse
    //fmt.Println("Resp: ", resp)
    if err := xml.NewDecoder(resp.Body).Decode(&defaultVPC); err != nil {
        fmt.Println("Error decoding response: ", err)
        return "", fmt.Errorf("failed to decode response: %w", err)
    }
    return defaultVPC.Vpcs[0].VpcId, nil


}
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