
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