package imaging

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	aws "nvms/lib/awspin"
	"nvms/models"
	"strings"
	"time"

	spinhttp "github.com/fermyon/spin-go-sdk/http"
)

// NewIMG creates a new IMG Client
func NewIMG(config aws.Config) (*Client, error) {
	u, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint: %w", err)
	}
	usePathStyle := strings.Contains(u.Host, "localhost") || strings.Contains(u.Host, "127.0.0.1")

		client := &Client{
			config:       config,
			endpointURL: u.String(),
			usePathStyle: usePathStyle,
		}

	return client, nil
}
func (c *Client) buildEndpoint() (string, error) {
     u, err := url.Parse(c.endpointURL)
    if err != nil {
        return "", fmt.Errorf("failed to parse endpoint: %w", err)
    }
    return u.String(), nil
}

func (c *Client) newRequest(ctx context.Context, method string, params map[string]string, body []byte, path string) (*http.Request, error) {
    endpoint, err := c.buildEndpoint()
    if err != nil {
        return nil, err
    }
    u, err := url.Parse(endpoint + path)
    if err != nil {
        return nil, err
    }
      
    req, err := http.NewRequestWithContext(ctx, method, u.String(),  bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

 // Set required headers
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("User-Agent", "byteport")

    // Add AWS v4 signature
    var awsDate aws.AwsDate
    awsDate.Time = time.Now()
    
    // Create canonical request
    payloadHash := aws.GetSHA256Hash(body)
    
    canonicalURI := path
    canonicalQueryString := ""
    canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\n",
        req.Header.Get("Content-Type"),
        u.Host)
    signedHeaders := "content-type;host"
    
    canonicalRequest := strings.Join([]string{
        method,
        canonicalURI,
        canonicalQueryString,
        canonicalHeaders,
        signedHeaders,
        payloadHash,
    }, "\n")

    // Create string to sign
    credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request",
        awsDate.GetDate(),
        c.config.Region,
        c.config.Service)

    stringToSign := strings.Join([]string{
        "AWS4-HMAC-SHA256",
        awsDate.GetTime(),
        credentialScope,
        aws.GetSHA256Hash([]byte(canonicalRequest)),
    }, "\n")

    // Calculate signature
    dateKey := aws.HmacSHA256([]byte("AWS4"+c.config.SecretAccessKey), []byte(awsDate.GetDate()))
    regionKey := aws.HmacSHA256(dateKey, []byte(c.config.Region))
    serviceKey := aws.HmacSHA256(regionKey, []byte(c.config.Service))
    signingKey := aws.HmacSHA256(serviceKey, []byte("aws4_request"))
    signature := hex.EncodeToString(aws.HmacSHA256(signingKey, []byte(stringToSign)))

    // Add authorization header
    authHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
        c.config.AccessKeyId,
        credentialScope,
        signedHeaders,
        signature)
    
    req.Header.Set("Authorization", authHeader)
    req.Header.Set("X-Amz-Date", awsDate.GetTime())
    
    if c.config.SessionToken != "" {
        req.Header.Set("X-Amz-Security-Token", c.config.SessionToken)
    }
    fmt.Println("Request created successfully: ", req)
    return req, nil
}
 
 
func (c *Client) do(req *http.Request) (*http.Response, error) {
    resp, err := spinhttp.Send(req)
    if err != nil {
		fmt.Println("Error sending request: ", err)
        return nil, fmt.Errorf("failed to send request: %w", err)
    }

    if resp.StatusCode != http.StatusOK {
		fmt.Println("Code: ", resp.StatusCode)
		fmt.Println("Response: ", resp)
        var errorResponse aws.ErrorResponse
        if err := xml.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			fmt.Println("Error parsing response: ", err)
            return nil, fmt.Errorf("failed to parse error response: %w", err)
        }
		fmt.Println("Error response: ", errorResponse)
        return nil, errorResponse
    }
	fmt.Println("Request sent successfully")
    return resp, nil
}

// Client provides an interface for interacting with the IMG API
type Client struct {
    config       aws.Config
    endpointURL string
    usePathStyle bool
}

// Instance represents an IMG instance
type Instance struct {
    InstanceId string `xml:"instanceId"`
    ImageId    string `xml:"imageId"`
    State      struct {
        Code int    `xml:"code"`
        Name string `xml:"name"`
    } `xml:"instanceState"`
    PrivateDnsName string `xml:"privateDnsName"`
    DnsName        string `xml:"dnsName"`
    Reason         string `xml:"reason"`
    KeyName        string `xml:"keyName"`
    AmiLaunchIndex int    `xml:"amiLaunchIndex"`
    ProductCodes   []string `xml:"productCodes"`
    InstanceType   string `xml:"instanceType"`
    LaunchTime     string `xml:"launchTime"`
    Placement      struct {
        AvailabilityZone string `xml:"availabilityZone"`
        GroupName        string `xml:"groupName"`
    } `xml:"placement"`
    Monitoring struct {
        State string `xml:"state"`
    } `xml:"monitoring"`
    SubnetId          string `xml:"subnetId"`
    VpcId             string `xml:"vpcId"`
    PrivateIpAddress  string `xml:"privateIpAddress"`
    SourceDestCheck   bool   `xml:"sourceDestCheck"`
    GroupSet          []struct {
        GroupId   string `xml:"groupId"`
        GroupName string `xml:"groupName"`
    } `xml:"groupSet>item"`
    Architecture      string `xml:"architecture"`
    RootDeviceType    string `xml:"rootDeviceType"`
    RootDeviceName    string `xml:"rootDeviceName"`
    BlockDeviceMapping []struct {
        DeviceName string `xml:"deviceName"`
        Ebs        struct {
            VolumeId            string `xml:"volumeId"`
            Status              string `xml:"status"`
            AttachTime          string `xml:"attachTime"`
            DeleteOnTermination bool   `xml:"deleteOnTermination"`
        } `xml:"ebs"`
    } `xml:"blockDeviceMapping>item"`
    VirtualizationType string `xml:"virtualizationType"`
    ClientToken        string `xml:"clientToken"`
    TagSet             []struct {
        Key   string `xml:"key"`
        Value string `xml:"value"`
    } `xml:"tagSet>item"`
    Hypervisor string `xml:"hypervisor"`
    NetworkInterfaceSet []struct {
        NetworkInterfaceId string `xml:"networkInterfaceId"`
        SubnetId           string `xml:"subnetId"`
        VpcId              string `xml:"vpcId"`
        Description        string `xml:"description"`
        OwnerId            string `xml:"ownerId"`
        Status             string `xml:"status"`
        MacAddress         string `xml:"macAddress"`
        PrivateIpAddress   string `xml:"privateIpAddress"`
        SourceDestCheck    bool   `xml:"sourceDestCheck"`
        GroupSet           []struct {
            GroupId   string `xml:"groupId"`
            GroupName string `xml:"groupName"`
        } `xml:"groupSet>item"`
        Attachment struct {
            AttachmentId         string `xml:"attachmentId"`
            DeviceIndex          int    `xml:"deviceIndex"`
            Status               string `xml:"status"`
            AttachTime           string `xml:"attachTime"`
            DeleteOnTermination  bool   `xml:"deleteOnTermination"`
        } `xml:"attachment"`
        PrivateIpAddressesSet []struct {
            PrivateIpAddress string `xml:"privateIpAddress"`
            Primary          bool   `xml:"primary"`
        } `xml:"privateIpAddressesSet>item"`
    } `xml:"networkInterfaceSet>item"`
    EbsOptimized bool `xml:"ebsOptimized"`

}
 
// RunInstancesResponse represents the response from RunInstances
type RunInstancesResponse struct {
    XMLName      xml.Name `xml:"RunInstancesResponse"`
    ReservationId string   `xml:"reservationId"`
    OwnerId       string   `xml:"ownerId"`
    Instances     []Instance `xml:"instancesSet>item"`
    
}

// DescribeInstancesResponse represents the response from DescribeInstances
type DescribeInstancesResponse struct {
    XMLName     xml.Name   `xml:"DescribeInstancesResponse"`
    Reservations []struct {
        Instances []Instance `xml:"instancesSet>item"`
    } `xml:"reservationSet>item"`
}
type CreatePipelineRequest struct {
   ClientToken string `json:"clientToken"`
   ContainerRecipeArn string `json:"containerRecipeArn"`
   Description string `json:"description"`
   DistributionConfigurationArn string `json:"distributionConfigurationArn"`
   EnhancedImageMetadataEnabled bool `json:"enhancedImageMetadataEnabled"`
   ExecutionRole string `json:"executionRole"`
   ImageRecipeArn string `json:"imageRecipeArn"`
   ImageScanningConfiguration struct {
      EcrConfiguration struct {
         ContainerTags []string `json:"containerTags"`
         RepositoryName string `json:"repositoryName"`
      } `json:"ecrConfiguration"`
      ImageScanningEnabled bool `json:"imageScanningEnabled"`
   } `json:"imageScanningConfiguration"`
   ImageTestsConfiguration struct {
      ImageTestsEnabled bool `json:"imageTestsEnabled"`
      TimeoutMinutes int `json:"timeoutMinutes"`
   } `json:"imageTestsConfiguration"`
   InfrastructureConfigurationArn string `json:"infrastructureConfigurationArn"`
   Name string `json:"name"`
   Schedule struct {
      PipelineExecutionStartCondition string `json:"pipelineExecutionStartCondition"`
      ScheduleExpression string `json:"scheduleExpression"`
      Timezone string `json:"timezone"`
   } `json:"schedule"`
   Status string `json:"status"`
   Tags map[string]string `json:"tags"`
   Workflows []struct {
      OnFailure string `json:"onFailure"`
      ParallelGroup string `json:"parallelGroup"`
      Parameters []struct {
         Name string `json:"name"`
         Value []string `json:"value"`
      } `json:"parameters"`
      WorkflowArn string `json:"workflowArn"`
   } `json:"workflows"`
}
 

type CreateInfrastructureConfigurationRequest struct {
   ClientToken string `json:"clientToken"`
   Description string `json:"description"`
   InstanceMetadataOptions struct {
      HttpPutResponseHopLimit int `json:"httpPutResponseHopLimit,omitempty"`
      HttpTokens string `json:"httpTokens,omitempty"`
   } `json:"instanceMetadataOptions"`
   InstanceProfileName string `json:"instanceProfileName"`
   InstanceTypes []string `json:"instanceTypes,omitempty"`
   KeyPair string `json:"keyPair,omitempty"`
   Logging struct {
      S3Logs struct {
         S3BucketName string `json:"s3BucketName,omitempty"`
         S3KeyPrefix string `json:"s3KeyPrefix,omitempty"`
      } `json:"s3Logs"`
   } `json:"logging,omitempty"`
   Name string `json:"name"`
   Placement struct {
      AvailabilityZone string `json:"availabilityZone,omitempty"`
      HostId string `json:"hostId,omitempty"`
      HostResourceGroupArn string `json:"hostResourceGroupArn,omitempty"`
      Tenancy string `json:"tenancy,omitempty"`
   } `json:"placement,omitempty"`
   ResourceTags map[string]string `json:"resourceTags,omitempty"`
   SecurityGroupIds []string `json:"securityGroupIds,omitempty"`
   SnsTopicArn string `json:"snsTopicArn,omitempty"`
   SubnetId string `json:"subnetId,omitempty"`
   Tags map[string]string `json:"tags,omitempty"`
   TerminateInstanceOnFailure bool `json:"terminateInstanceOnFailure"`

}
 
type ExecuteImgPipelineRequest struct {
   ClientToken string `json:"clientToken"`
   ImagePipelineArn string `json:"imagePipelineArn"`
}
type CreatePipelineResponse struct {
	ClientToken string `json:"clientToken"`
	ImagePipelineArn string `json:"imagePipelineArn"`
	RequestId string `json:"requestId"`
}
type DeletePipelineResponse struct {
   ImagePipelineArn string `json:"imagePipelineArn"`
   RequestId string `json:"requestId"`
}

type ExecuteImgPipelineResponse struct {
   ClientToken string   `json:"clientToken"`
   ImageBuildVersionArn string `json:"imageBuildVersionArn"`
   RequestId string `json:"requestId"`
}



 
type CreateInfrastructureConfigurationResponse struct{
   ClientToken string `json:"clientToken"`
   InfrastructureConfigurationArn string `json:"infrastructureConfigurationArn"`
   RequestId string `json:"requestId"`

}
 
type DeleteInfrastructureConfigurationResponse struct{
   InfrastructureConfigurationArn string `json:"infrastructureConfigurationArn"`
   RequestId string `json:"requestId"`
}
type ImageRecipe struct {
    ParentImage     string           // Base image
    Components      []models.ImageComponent // From buildpack
    WorkingDir     string           // Service path
    Tags           map[string]string
}
 

type CreateImageRecipeResponse struct {
   ClientToken string `json:"clientToken"`
   ImageRecipeArn string `json:"imageRecipeArn"`
   RequestId string `json:"requestId"`

}
 type ContainerRepo struct{
   RepositoryName string `json:"repositoryName"`
   Service string `json:"service"`
 }
type CreateContainerRecipeRequest struct {
   ClientToken string `json:"clientToken"`
   Components []ComponentsConfig `json:"components"`
   ContainerType string `json:"containerType"`
   Description string `json:"description,omitempty"`
   DockerfileTemplateData string `json:"dockerfileTemplateData,omitempty"`
   DockerfileTemplateUri string `json:"dockerfileTemplateUri,omitempty"`
   ImageOsVersionOverride string `json:"imageOsVersionOverride,omitempty"`
   InstanceConfiguration struct {
      BlockDeviceMappings []struct {
         DeviceName string `json:"deviceName,omitempty"`
         Ebs struct {
            DeleteOnTermination bool `json:"deleteOnTermination,omitempty"`
            Encrypted bool `json:"encrypted,omitempty"`
            Iops int `json:"iops,omitempty"`
            KmsKeyId string `json:"kmsKeyId,omitempty"`
            SnapshotId string `json:"snapshotId,omitempty"`
            Throughput int `json:"throughput,omitempty"`
            VolumeSize int `json:"volumeSize,omitempty"`
            VolumeType string `json:"volumeType,omitempty"`
         } `json:"ebs,omitempty"`
         NoDevice string `json:"noDevice,omitempty"`
         VirtualName string `json:"virtualName,omitempty"`
      } `json:"blockDeviceMappings,omitempty"`
      Image string `json:"image,omitempty"`
   } `json:"instanceConfiguration,omitempty"`
   KmsKeyId string `json:"kmsKeyId,omitempty"`
   Name string `json:"name"`
   ParentImage string `json:"parentImage"`
   PlatformOverride string `json:"platformOverride,omitempty"`
   SemanticVersion string `json:"semanticVersion"`
   Tags map[string]string `json:"tags,omitempty"`
   TargetRepository ContainerRepo  `json:"targetRepository"`
   WorkingDirectory string `json:"workingDirectory,omitempty"`

}
type CreateContainerRecipeResponse struct {
   ClientToken string `json:"clientToken"`
   ContainerRecipeArn string `json:"containerRecipeArn"`
   RequestId string `json:"requestId"`
}
type ComponentsConfigParams struct{
    Name string `json:"name"`
    Value []string `json:"value"`
}
type ComponentsConfig struct {
   ComponentArn string `json:"componentArn"`
   Parameters []ComponentsConfigParams   `json:"parameters,omitempty"`}
type DeleteContainerRecipeResponse struct {
	ContainerRecipeArn string `json:"containerRecipeArn"`
	RequestId string `json:"requestId"`
}
type CreateImageRecipeRequest struct {
   AdditionalInstanceConfiguration struct{
      SystemsManagerAgent struct{
         UninstallAfterBuild bool `json:"uninstallAfterBuild"`
      } `json:"systemsManagerAgent"`
      UserDataOverride string `json:"userDataOverride"`
      
   }
   BlockDeviceMappings struct{
      DeviceName string `json:"deviceName"`
      Ebs struct {
         DeleteOnTermination bool `json:"deleteOnTermination"`
         Encrypted bool `json:"encrypted,omitempty"`
         Iops int `json:"iops,omitempty"`
         KmsKeyId string `json:"kmsKeyId,omitempty"`
         SnapshotId string `json:"snapshotId,omitempty"`
         Throughput int `json:"throughput,omitempty"`
         VolumeSize int `json:"volumeSize,omitempty"`
         VolumeType string `json:"volumeType,omitempty"`

      }
      NoDevice string `json:"noDevice,omitempty"`
      VirtualName string `json:"virtualName,omitempty"`
   } `json:"blockDeviceMappings,omitempty"`
 
   ClientToken string `json:"clientToken"`
   Components []ComponentsConfig `json:"components"`
   Description string `json:"description,omitempty"`
   Name string `json:"name"`
   ParentImage string `json:"parentImage"`
   SemanticVersion string `json:"semanticVersion"`
   Tags map[string]string `json:"tags,omitempty"`
   WorkingDirectory string `json:"workingDirectory"`

}
type DeleteImageRecipeResponse struct {
   ImageRecipeArn string `json:"imageRecipeArn"`
   RequestId string `json:"requestId"`
}
 
type CreateImageComponentRequest struct {
   ChangeDescription string `json:"changeDescription,omitempty"`
   ClientToken string `json:"clientToken"`
   Data string `json:"data"`
   Description string `json:"description,omitempty"`
   KmsKeyId string `json:"kmsKeyId,omitempty"`
   Name string `json:"name"`
   Platform string `json:"platform"`
   SemanticVersion string `json:"semanticVersion"`
   SupportedOsVersions []string `json:"supportedOsVersions,omitempty"`
   Tags map[string]string `json:"tags,omitempty"`
   Uri string `json:"uri,omitempty"`

}
 
type CreateImageComponentResponse struct {
   ClientToken string `json:"clientToken"`
   ComponentBuildVersionArn string `json:"componentBuildVersionArn"`
   RequestId string `json:"requestId"`
}
//type DeleteImageComponentResponse struct {}
