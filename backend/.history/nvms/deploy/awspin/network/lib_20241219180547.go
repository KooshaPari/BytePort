package network

import aws "nvms/deploy/awspin"
type Client struct {
    config       aws.Config
    endpointURL string
}

/* ALB RESPONSE: 
<CreateLoadBalancerResponse xmlns="http://elasticloadbalancing.amazonaws.com/
 doc/2015-12-01/"> 
  <CreateLoadBalancerResult> 
    <LoadBalancers>  
      <member>  
        <LoadBalancerArn>arn:aws:elasticloadbalancing:us
west-2:123456789012:loadbalancer/app/my-internal-load-balancer/50dc6c495c0c9188</
 LoadBalancerArn>  
        <Scheme>internet-facing</Scheme>  
        <LoadBalancerName>my-load-balancer</LoadBalancerName>  
        <VpcId>vpc-3ac0fb5f</VpcId>  
        <CanonicalHostedZoneId>Z2P70J7EXAMPLE</CanonicalHostedZoneId>  
        <CreatedTime>2016-03-25T21:29:48.850Z</CreatedTime>  
        <AvailabilityZones>  
          <member>  
            <SubnetId>subnet-8360a9e7</SubnetId>  
            <ZoneName>us-west-2a</ZoneName>  
          </member>  
          <member>  
            <SubnetId>subnet-b7d581c0</SubnetId>  
            <ZoneName>us-west-2b</ZoneName>  
          </member>  
        </AvailabilityZones>  
        <SecurityGroups>  
          <member>sg-5943793c</member>  
        </SecurityGroups>  
        <DNSName>my-load-balancer-424835706.us-west-2.elb.amazonaws.com</DNSName>  
        <State>  
          <Code>provisioning</Code>  
        </State>  
        <Type>application</Type>  
      </member>  
Examples
 API Version 2015-12-01 30
Elastic Load Balancing API Reference
    </LoadBalancers>  
  </CreateLoadBalancerResult>  
  <ResponseMetadata>  
    <RequestId>32d531b2-f2d0-11e5-9192-3fff33344cfa</RequestId>  
  </ResponseMetadata>
 </CreateLoadBalancerResponse>*/
 type CreateLoadBalancerResponse struct {
	LoadBalancers struct {
		Member struct {
			LoadBalancerArn string `xml:"LoadBalancerArn"`
			Scheme string `xml:"Scheme"`
			LoadBalancerName string `xml:"LoadBalancerName"`
			VpcId string `xml:"VpcId"`
			CanonicalHostedZoneId string `xml:"CanonicalHostedZoneId"`
			CreatedTime string `xml:"CreatedTime"`
			AvailabilityZones struct {
				Member struct {
					SubnetId string `xml:"SubnetId"`
					