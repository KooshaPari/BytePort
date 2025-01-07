package ec2

import "encoding/xml"

/*
 <CreateInstanceProfileResponse xmlns="https://iam.amazonaws.com/doc/2010-05-08/">
  <CreateInstanceProfileResult>
    <InstanceProfile>
      <InstanceProfileId>AIPAD5ARO2C5EXAMPLE3G</InstanceProfileId>
      <Roles/>
      <InstanceProfileName>Webserver</InstanceProfileName>
      <Path>/application_abc/component_xyz/</Path>
      <Arn>arn:aws:iam::123456789012:instance-profile/application_abc/component_xyz/Webserver</Arn>
      <CreateDate>2012-05-09T16:11:10.222Z</CreateDate>
    </InstanceProfile>
  </CreateInstanceProfileResult>
  <ResponseMetadata>
    <RequestId>974142ee-99f1-11e1-a4c3-27EXAMPLE804</RequestId>
  </ResponseMetadata>
</CreateInstanceProfileResponse>*/
type CreateInstanceProfileResponse struct {
	XMLName xml.Name `xml:"https://iam.amazonaws.com/doc/2010-05-08/ CreateInstanceProfileResponse"`
	CreateInstanceProfileResult struct {
		InstanceProfile struct {
			InstanceProfileId string `xml:"InstanceProfileId"`
			Roles []string `xml:"Roles"`
			InstanceProfileName string `xml:"InstanceProfileName"`
			Path string `xml:"Path"`
			Arn string `xml:"Arn"`
			CreateDate string `xml:"CreateDate"`
		} `xml:"InstanceProfile"`
	} `xml:"CreateInstanceProfileResult"`
	ResponseMetadata struct {
		RequestId string `xml:"RequestId"`
	} `xml:"ResponseMetadata"`
}

	type DeleteInstanceProfileResponse struct {
		XMLName xml.Name `xml:"https://iam.amazonaws.com/doc/2010-05-08/ DeleteInstanceProfileResponse"`
		ResponseMetadata struct {
		RequestId string `xml:"RequestId"`
	} `xml:"ResponseMetadata"`
	}
type Get