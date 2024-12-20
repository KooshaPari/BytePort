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
			UnhealthyThresholdCount string `xml:"UnhealthyThresholdCount"`
		} `xml:"member"`
	} `xml:"TargetGroups"`
	ResponseMetadata struct {
		RequestId string `xml:"RequestId"`
	} `xml:"ResponseMetadata"`
}
/*
<CreateRuleResponse xmlns="http://elasticloadbalancing.amazonaws.com/doc/2015-12-01/">
  <CreateRuleResult> 
    <Rules> 
      <member> 
        <IsDefault>false</IsDefault> 
        <Conditions> 
          <member> 
            <Field>path-pattern</Field> 
            <Values> 
              <member>/img/*</member> 
            </Values> 
          </member> 
        </Conditions> 
        <Priority>10</Priority> 
        <Actions> 
          <member> 
            <Type>forward</Type> 
            <TargetGroupArn>arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/my-targets/73e2d6bc24d8a067</TargetGroupArn> 
          </member> 
        </Actions> 
        <RuleArn>arn:aws:elasticloadbalancing:us-west-2:123456789012:listener-rule/app/my-load-balancer/50dc6c495c0c9188/f2f7dc8efc522ab2/9683b2d02a6cabee</RuleArn> 
      </member> 
    </Rules> 
  </CreateRuleResult> 
  <ResponseMetadata> 
    <RequestId>c5478c83-f397-11e5-bb98-57195a6eb84a</RequestId> 
  </ResponseMetadata>
</CreateRuleResponse>
*/
type CreateRuleResponse struct {
	Rules struct {
		 
}