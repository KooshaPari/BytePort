package network

import aws "nvms/deploy/awspin"
type Client struct {
    config       aws.Config
    endpointURL string
}

 
 type CreateLoadBalancerResponse struct {
	LoadBalancers struct {
		Member []struct {
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
	Listeners  []struct {
		Member  struct {
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
	TargetGroups  []struct {
		Member  struct {
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
 
type CreateRuleResponse struct {
	Rules []struct {
		Member  struct {
			IsDefault string `xml:"IsDefault"`
			Conditions struct {
				Member struct {
					Field string `xml:"Field"`
					Values struct {
						Member string `xml:"member"`
					} `xml:"Values"`
				} `xml:"member"`
			} `xml:"Conditions"`
			Priority string `xml:"Priority"`
			Actions struct {
				Member struct {
					Type string `xml:"Type"`
					TargetGroupArn string `xml:"TargetGroupArn"`
				} `xml:"Actions"`
			} `xml:"Actions"`
			RuleArn string `xml:"RuleArn"`
		} `xml:"member"`
	} `xml:"Rules"` 
}