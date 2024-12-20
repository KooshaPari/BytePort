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
 
type CreateListenerResponse struct {
	Listeners struct {
		Member struct {
			LoadBalancerArn string `xml:"LoadBalancerArn"`
			Protocol string `xml:"Protocol"`
			Port string `xml:"Port"`
			ListenerArn string `xml:"ListenerArn"`
			DefaultActions struct {
				Member struct {
					Type string `xml:"Type"`
					TargetGroupArn string `xml:"TargetGroupArn"`
				} `xml:"member"`
			} `xml:"DefaultActions"`
		} `xml:"member"`
	} `xml:"Listeners"`
	ResponseMetadata struct {
		RequestId string `xml:"RequestId"`
	} `xml:"ResponseMetadata"` 


}
/*
<CreateTargetGroupResponse xmlns="http://elasticloadbalancing.amazonaws.com/doc/2015-12-01/">
  <CreateTargetGroupResult> 
    <TargetGroups> 
      <member> 
        <TargetGroupArn>arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/my-targets/73e2d6bc24d8a067</TargetGroupArn> 
        <HealthCheckTimeoutSeconds>5</HealthCheckTimeoutSeconds> 
        <HealthCheckPort>traffic-port</HealthCheckPort> 
        <Matcher>
          <HttpCode>200</HttpCode> 
        </Matcher> 
        <TargetGroupName>my-targets</TargetGroupName> 
        <HealthCheckProtocol>HTTP</HealthCheckProtocol> 
        <HealthCheckPath>/</HealthCheckPath> 
        <Protocol>HTTP</Protocol> 
        <Port>80</Port> 
        <VpcId>vpc-3ac0fb5f</VpcId> 
        <HealthyThresholdCount>5</HealthyThresholdCount> 
        <HealthCheckIntervalSeconds>30</HealthCheckIntervalSeconds> 
        <UnhealthyThresholdCount>2</UnhealthyThresholdCount> 
      </member> 
    </TargetGroups> 
  </CreateTargetGroupResult> 
  <ResponseMetadata> 
    <RequestId>b83fe90e-f2d5-11e5-b95d-3b2c1831fc26</RequestId> 
  </ResponseMetadata>
</CreateTargetGroupResponse>
*/
type CreateTargetGroupResponse struct {
	TargetGroups struct {
		Member struct {
			TargetGroupArn string `xml:"TargetGroupArn"`
			HealthCheckTimeoutSeconds string `xml:"HealthCheckTimeoutSeconds"`
			HealthCheckPort string `xml:"HealthCheckPort"`
			Matcher struct {
				HttpCode string `xml:"HttpCode"`
			} `xml:"Matcher"`
			TargetGroupName string `xml:"TargetGroupName"`
			HealthCheckProtocol string `xml:"HealthCheckProtocol"`
			HealthCheckPath string `xml:"HealthCheckPath"`
			Protocol string `xml:"Protocol"`
			Port string `xml:"Port"`
			VpcId string `xml:"VpcId"`
			HealthyThresholdCount string `xml:"HealthyThresholdCount"`
			Unh
}