package network

import aws "nvms/deploy/awspin"

type Client struct {
    config      aws.Config
    endpointURL string
}

type CreateLoadBalancerResponse struct {
    LoadBalancers struct {
        Items []struct {
            LoadBalancerArn       string `xml:"LoadBalancerArn"`
            Scheme                string `xml:"Scheme"`
            LoadBalancerName      string `xml:"LoadBalancerName"`
            VpcId                 string `xml:"VpcId"`
            CanonicalHostedZoneId string `xml:"CanonicalHostedZoneId"`
            CreatedTime           string `xml:"CreatedTime"`
            AvailabilityZones     struct {
                Items []struct {
                    SubnetId string `xml:"SubnetId"`
                    ZoneName string `xml:"ZoneName"`
                } `xml:"item"`
            } `xml:"AvailabilityZones"`
            SecurityGroups struct {
                Items []string `xml:"item"`
            } `xml:"SecurityGroups"`
            DNSName string `xml:"DNSName"`
            State   struct {
                Code string `xml:"Code"`
            } `xml:"State"`
            Type string `xml:"Type"`
        } `xml:"item"`
    } `xml:"LoadBalancers"`
    ResponseMetadata struct {
        RequestId string `xml:"RequestId"`
    } `xml:"ResponseMetadata"`
}

type CreateListenerResponse struct {
    Listeners struct {
        Items []struct {
            LoadBalancerArn string `xml:"LoadBalancerArn"`
            Protocol        string `xml:"Protocol"`
            Port            string `xml:"Port"`
            ListenerArn     string `xml:"ListenerArn"`
            DefaultActions  struct {
                Items []struct {
                    Type          string `xml:"Type"`
                    TargetGroupArn string `xml:"TargetGroupArn"`
                } `xml:"item"`
            } `xml:"DefaultActions"`
        } `xml:"item"`
    } `xml:"Listeners"`
    ResponseMetadata struct {
        RequestId string `xml:"RequestId"`
    } `xml:"ResponseMetadata"`
}

type CreateTargetGroupResponse struct {
    TargetGroups struct {
        Items []struct {
            TargetGroupArn             string `xml:"TargetGroupArn"`
            HealthCheckTimeoutSeconds  string `xml:"HealthCheckTimeoutSeconds"`
            HealthCheckPort            string `xml:"HealthCheckPort"`
            Matcher                    struct {
                HttpCode string `xml:"HttpCode"`
            } `xml:"Matcher"`
            TargetGroupName            string `xml:"TargetGroupName"`
            HealthCheckProtocol        string `xml:"HealthCheckProtocol"`
            HealthCheckPath            string `xml:"HealthCheckPath"`
            Protocol                   string `xml:"Protocol"`
            Port                       string `xml:"Port"`
            VpcId                      string `xml:"VpcId"`
            HealthyThresholdCount      string `xml:"HealthyThresholdCount"`
            UnhealthyThresholdCount    string `xml:"UnhealthyThresholdCount"`
        } `xml:"item"`
    } `xml:"TargetGroups"`
    ResponseMetadata struct {
        RequestId string `xml:"RequestId"`
    } `xml:"ResponseMetadata"`
}

type CreateRuleResponse struct {
    Rules struct {
        Items []struct {
            IsDefault   string `xml:"IsDefault"`
            Conditions  struct {
                Items []struct {
                    Field  string `xml:"Field"`
                    Values struct {
                        Items []string `xml:"item"`
                    } `xml:"Values"`
                } `xml:"item"`
            } `xml:"Conditions"`
            Priority string `xml:"Priority"`
            Actions  struct {
                Items []struct {
                    Type          string `xml:"Type"`
                    TargetGroupArn string `xml:"TargetGroupArn"`
                } `xml:"item"`
            } `xml:"Actions"`
            RuleArn string `xml:"RuleArn"`
        } `xml:"item"`
    } `xml:"Rules"`
    ResponseMetadata struct {
        RequestId string `xml:"RequestId"`
    } `xml:"ResponseMetadata"`
}