package ec2

import "encoding/xml"
type DescribeVpcsResponse struct {
    XMLName     xml.Name   `xml:"DescribeVpcsResponse"`
    Vpcs []struct {
        VpcId string `xml:"vpcId"`
        OwnerId string `xml:"ownerId"`
        State string `xml:"state"`
        CidrBlock string `xml:"cidrBlock"`
        CidrBlockAssociationSet []struct {
            CidrBlock string `xml:"cidrBlock"`
            AssociationId string `xml:"associationId"`
            CidrBlockState struct {
                State string `xml:"state"`
            } `xml:"cidrBlockState"`
        } `xml:"cidrBlockAssociationSet>item"`
        DhcpOptionsId string `xml:"dhcpOptionsId"`
        TagSet []struct {
            Key string `xml:"key"`
            Value string `xml:"value"`
        } `xml:"tagSet>item"`
        InstanceTenancy string `xml:"instanceTenancy"`
        IsDefault bool `xml:"isDefault"`
    } `xml:"vpcSet>item"`
}
 