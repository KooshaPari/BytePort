package ec2

import "encoding/xml"
 
type CreateInstanceProfileResponse struct {
	XMLName xml.Name `xml:"https://iam.amazonaws.com/doc/2010-05-08/ CreateInstanceProfileResponse"`
	InstanceProfile struct {
		InstanceProfileId string `xml:"InstanceProfileId"`
		Roles []struct {
			RoleName string `xml:"RoleName"`
			RoleId string `xml:"RoleId"`
			Arn string `xml:"Arn"`
		} `xml:"Roles>member"`
		InstanceProfileName string `xml:"InstanceProfileName"`
		Path string `xml:"Path"`
		Arn string `xml:"Arn"`
		CreateDate string `xml:"CreateDate"`
	} `xml:"CreateInstanceProfileResult>InstanceProfile"`
	ResponseMetadata struct {
		RequestId string `xml:"RequestId"`
	} `xml:"ResponseMetadata"`
	 
	} 
	type DeleteInstanceProfileResponse struct {
		XMLName xml.Name `xml:"https://iam.amazonaws.com/doc/2010-05-08/ DeleteInstanceProfileResponse"`
		