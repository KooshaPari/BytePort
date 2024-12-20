package vpc

import (
	aws "nvms/deploy/awspin"
)

// Client provides an interface for interacting with the EC2 API
type Client struct {
    config       aws.Config
    endpointURL string
}

/*
<CreateVpcResponse xmlns="http://ec2.amazonaws.com/doc/2016-11-15/">
    <requestId>b1a2b2b5-5806-4e24-824b-0c8996c608c1</requestId>
    <vpc>
        <vpcId>vpc-03914afb3ed6c7632</vpcId>
        <ownerId>111122223333</ownerId>
        <state>pending</state>
        <cidrBlock>10.0.0.0/16</cidrBlock>
        <cidrBlockAssociationSet>
            <item>
                <cidrBlock>10.0.0.0/16</cidrBlock>
                <associationId>vpc-cidr-assoc-03ca48bbbeEXAMPLE</associationId>
                <cidrBlockState>
                    <state>associated</state>
                </cidrBlockState>
            </item>
        </cidrBlockAssociationSet>
        <ipv6CidrBlockAssociationSet>
            <item>
                <ipv6CidrBlock></ipv6CidrBlock>
                <associationId>vpc-cidr-assoc-0bd6cc7621EXAMPLE</associationId>
                <ipv6CidrBlockState>
                    <state>associating</state>
                </ipv6CidrBlockState>
                <Ipv6CidrBlockNetworkBorderGroup>us-west-2-lax-1</Ipv6CidrBlockNetworkBorderGroup>
            </item>
        </ipv6CidrBlockAssociationSet>
        <dhcpOptionsId>dopt-19edf471</dhcpOptionsId>
        <tagSet/>
        <instanceTenancy>default</instanceTenancy>
        <isDefault>false</isDefault>
        <availabilityZone>us-west-2-lax-1a</availabilityZone>
    </vpc>
</CreateVpcResponse>
*/
type CreateVpcResponse struct {
    XMLName   xml.Name `xml:"CreateVpcResponse"`
    RequestId string `xml:"requestId"`
    Vpc       struct {
        VpcId string `xml:"vpcId"`
        OwnerId string `xml:"ownerId"`
        State string `xml:"state"`
        CidrBlock string `xml:"cidrBlock"`
        CidrBlockAssociationSet struct {
            Item struct {
                CidrBlock string `xml:"cidrBlock"`
                AssociationId string `xml:"associationId"`
                CidrBlockState struct {
                    State string `xml:"state"`
                } `xml:"cidrBlockState"`
            } `xml:"item"`
        } `xml:"cidrBlockAssociationSet"`
        Ipv6CidrBlockAssociationSet struct {
            Item struct {
                Ipv6CidrBlock string `xml:"ipv6CidrBlock"`
                AssociationId string `xml:"associationId"`
                Ipv6CidrBlockState struct {
                    State string `xml:"state"`
                } `xml:"ipv6CidrBlockState"`
                Ipv6CidrBlockNetworkBorderGroup string `xml:"Ipv6CidrBlockNetworkBorderGroup"`
            } `xml:"item"`
        } `xml:"ipv6CidrBlockAssociationSet"`
        DhcpOptionsId string `xml:"dhcpOptionsId"`
        TagSet struct {
        } `xml:"tagSet"`
        InstanceTenancy string `xml:"instanceTenancy"`
        IsDefault bool `xml:"isDefault"`
        AvailabilityZone string `xml:"availabilityZone"`
    } `xml:"vpc"`
    