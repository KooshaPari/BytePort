package ec2

import "encoding/xml"
type DescribeSecurityGroupsResponse struct {
    XMLName xml.Name `xml:"DescribeSecurityGroupsResponse"`
    SecurityGroupInfo struct {
        Item struct {
            OwnerId string `xml:"ownerId"`
            GroupId string `xml:"groupId"`
            GroupName string `xml:"groupName"`
            GroupDescription string `xml:"groupDescription"`
            VpcId string `xml:"vpcId"`
            IpPermissions []struct {
                Item struct {
                    IpProtocol string `xml:"ipProtocol"`
                    FromPort int `xml:"fromPort"`
                    ToPort int `xml:"toPort"`
                    Groups []struct {
                        Item struct {
                            SecurityGroupRuleId string `xml:"securityGroupRuleId"`
                            UserId string `xml:"userId"`
                            GroupId string `xml:"groupId"`
                            VpcId string `xml:"vpcId"`
                            VpcPeeringConnectionId string `xml:"vpcPeeringConnectionId"`
                            PeeringStatus string `xml:"peeringStatus"`
                        } `xml:"item"`
                    } `xml:"groups"`
                    IpRanges []struct {
                        Item struct {
                            CidrIp string `xml:"cidrIp"`
                        } `xml:"item"`
                    } `xml:"ipRanges"`
                    PrefixListIds []struct {
                        Item struct {
                            PrefixListId string `xml:"prefixListId"`
                        } `xml:"item"`
                    } `xml:"prefixListIds"`
                } `xml:"item"`
            } `xml:"ipPermissions"`
            IpPermissionsEgress []struct {
                Item struct {
                    IpProtocol string `xml:"ipProtocol"`
                    Groups []struct {
                        Item struct {
                            SecurityGroupRuleId string `xml:"securityGroupRuleId"`
                            UserId string `xml:"userId"`
                            GroupId string `xml:"groupId"`
                            VpcId string `xml:"vpcId"`
                            VpcPeeringConnectionId string `xml:"vpcPeeringConnectionId"`
                            PeeringStatus string `xml:"peeringStatus"`
                        } `xml:"item"`
                    } `xml:"groups"`
                    IpRanges []struct {
                        Item struct {
                            CidrIp string `xml:"cidrIp"`
                        } `xml:"item"`
                    } `xml:"ipRanges"`
                    PrefixListIds []struct {
                        Item struct {
                            PrefixListId string `xml:"prefixListId"`
                        } `xml:"item"`
                    } `xml:"prefixListIds"`
                } `xml:"item"`
            } `xml:"ipPermissionsEgress"`
        } `xml:"item"`
    } `xml:"securityGroupInfo"`
    

} 