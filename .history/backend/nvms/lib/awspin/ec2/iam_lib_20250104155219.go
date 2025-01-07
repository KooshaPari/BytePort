package ec2
import ("xml")
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
</CreateInstanceProfileResponse>
*/
type CreateInstanceProfileResponse struct {
	XML.Name xml.Name `xml:"https://iam.amazonaws.com/doc/2010-05-08/ CreateInstanceProfileResponse"`
	InstanceProfile struct {
		InstanceProfileId string `xml:"InstanceProfileId"`
		Roles []struct {
			RoleName string `xml:"RoleName"`
			RoleId string `xml:"RoleId"`
			Arn string `xml:"Arn"`
			
	}
}