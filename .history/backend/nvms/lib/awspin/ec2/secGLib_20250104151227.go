
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
 type DescribeSubnetsResponse struct {
    XMLName   xml.Name `xml:"DescribeSubnetsResponse"`
    SubnetSet []Subnet `xml:"subnetSet>item"`
}

type Subnet struct {
    SubnetId                     string `xml:"subnetId"`
    SubnetArn                    string `xml:"subnetArn"`
    State                        string `xml:"state"`
    OwnerId                      string `xml:"ownerId"`
    VpcId                        string `xml:"vpcId"`
    CidrBlock                    string `xml:"cidrBlock"`
    Ipv6CidrBlockAssociationSet  []Ipv6CidrBlockAssociation `xml:"ipv6CidrBlockAssociationSet>item"`
    AvailableIpAddressCount      int    `xml:"availableIpAddressCount"`
    AvailabilityZone             string `xml:"availabilityZone"`
    AvailabilityZoneId           string `xml:"availabilityZoneId"`
    DefaultForAz                 bool   `xml:"defaultForAz"`
    MapPublicIpOnLaunch          bool   `xml:"mapPublicIpOnLaunch"`
    AssignIpv6AddressOnCreation  bool   `xml:"assignIpv6AddressOnCreation"`
}

type Ipv6CidrBlockAssociation struct {
    Ipv6CidrBlock       string `xml:"ipv6CidrBlock"`
    AssociationId       string `xml:"associationId"`
    Ipv6CidrBlockState  struct {
        State string `xml:"state"`
    } `xml:"ipv6CidrBlockState"`
}