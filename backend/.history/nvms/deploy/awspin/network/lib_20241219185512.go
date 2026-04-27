package network

import aws "nvms/deploy/awspin"
type Client struct {
    config       aws.Config
    endpointURL string
}

 
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
					ZoneName string `xml:"ZoneName"`
				} `xml:"member"`
			} `xml:"AvailabilityZones"`
			SecurityGroups struct {
				Member string `xml:"member"`
			} `xml:"SecurityGroups"`
			DNSName string `xml:"DNSName"`
			State struct {
				Code string `xml:"Code"`
			} `xml:"State"`
			Type string `xml:"Type"`
		} `xml:"member"`
	} `xml:"LoadBalancers"`
	ResponseMetadata struct {
		RequestId string `xml:"RequestId"`
	} `xml:"ResponseMetadata"`
}
/*
<CreateListenerResponse xmlns="http://elasticloadbalancing.amazonaws.com/doc/2015-12-01/">
  <CreateListenerResult> 
    <Listeners> 
      <member> 
        <LoadBalancerArn>arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/app/my-load-balancer/50dc6c495c0c9188</LoadBalancerArn> 
        <Protocol>HTTP</Protocol> 
        <Port>80</Port> 
        <ListenerArn>arn:aws:elasticloadbalancing:us-west-2:123456789012:listener/app/my-load-balancer/50dc6c495c0c9188/f2f7dc8efc522ab2</ListenerArn> 
        <DefaultActions> 
          <member> 
            <Type>forward</Type> 
            <TargetGroupArn>arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/my-targets/73e2d6bc24d8a067</TargetGroupArn> 
          </member> 
        </DefaultActions> 
      </member> 
    </Listeners> 
  </CreateListenerResult> 
  <ResponseMetadata> 
    <RequestId>883c84bb-f387-11e5-ae48-cff02092876b</RequestId> 
  </ResponseMetadata> 
</CreateListenerResponse>*/
type CreateListenerResponse struct {

}