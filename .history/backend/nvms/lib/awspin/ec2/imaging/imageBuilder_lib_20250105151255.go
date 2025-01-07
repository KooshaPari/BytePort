package imaging

import (
	"context"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	aws "nvms/lib/awspin"
	"nvms/models"
	"sort"
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

func (c *Client) newRequest(ctx context.Context, method string, params map[string]string, body []byte) (*http.Request, error) {
    endpoint, err := c.buildEndpoint()
    if err != nil {
        return nil, err
    }
    u, err := url.Parse(endpoint)
    if err != nil {
        return nil, err
    }
     

    // Create request with minimal headers
    req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("host", u.Host)
    req.Header.Set("user-agent", "byteport")

    //fmt.Printf("Request headers: %+v\n", req.Header)
    return req, nil
}

// Helper function to create canonical query string
func GetCanonicalQueryString(params map[string]string) string {
    // Get sorted list of parameter names
    paramNames := make([]string, 0, len(params))
    for name := range params {
        paramNames = append(paramNames, name)
    }
    sort.Strings(paramNames)

    // Build canonical query string
    pairs := make([]string, 0, len(params))
    for _, name := range paramNames {
        pairs = append(pairs, fmt.Sprintf("%s=%s",
            url.QueryEscape(name),
            url.QueryEscape(params[name]),
        ))
    }

    return strings.Join(pairs, "&")
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
      HttpPutResponseHopLimit int `json:"httpPutResponseHopLimit"`
      HttpTokens string `json:"httpTokens"`
   } `json:"instanceMetadataOptions"`
   InstanceProfileName string `json:"instanceProfileName"`
   InstanceTypes []string `json:"instanceTypes"`
   KeyPair string `json:"keyPair"`
   Logging struct {
      S3Logs struct {
         S3BucketName string `json:"s3BucketName"`
         S3KeyPrefix string `json:"s3KeyPrefix"`
      } `json:"s3Logs"`
   } `json:"logging"`
   Name string `json:"name"`
   Placement struct {
      AvailabilityZone string `json:"availabilityZone"`
      HostId string `json:"hostId"`
      HostResourceGroupArn string `json:"hostResourceGroupArn"`
      Tenancy string `json:"tenancy"`
   } `json:"placement"`
   ResourceTags map[string]string `json:"resourceTags"`
   SecurityGroupIds []string `json:"securityGroupIds"`
   SnsTopicArn string `json:"snsTopicArn"`
   SubnetId string `json:"subnetId"`
   Tags map[string]string `json:"tags"`
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
   Description string `json:"description"`
   DockerfileTemplateData string `json:"dockerfileTemplateData"`
   DockerfileTemplateUri string `json:"dockerfileTemplateUri"`
   ImageOsVersionOverride string `json:"imageOsVersionOverride"`
   InstanceConfiguration struct {
      BlockDeviceMappings []struct {
         DeviceName string `json:"deviceName"`
         Ebs struct {
            DeleteOnTermination bool `json:"deleteOnTermination"`
            Encrypted bool `json:"encrypted"`
            Iops int `json:"iops"`
            KmsKeyId string `json:"kmsKeyId"`
            SnapshotId string `json:"snapshotId"`
            Throughput int `json:"throughput"`
            VolumeSize int `json:"volumeSize"`
            VolumeType string `json:"volumeType"`
         } `json:"ebs"`
         NoDevice string `json:"noDevice"`
         VirtualName string `json:"virtualName"`
      } `json:"blockDeviceMappings"`
      Image string `json:"image"`
   } `json:"instanceConfiguration"`
   KmsKeyId string `json:"kmsKeyId"`
   Name string `json:"name"`
   ParentImage string `json:"parentImage"`
   PlatformOverride string `json:"platformOverride"`
   SemanticVersion string `json:"semanticVersion"`
   Tags map[string]string `json:"tags"`
   TargetRepository ContainerRepo  `json:"targetRepository"`
   WorkingDirectory string `json:"workingDirectory"`

}
type CreateContainerRecipeResponse struct {
   ClientToken string `json:"clientToken"`
   ContainerRecipeArn string `json:"containerRecipeArn"`
   RequestId string `json:"requestId"`
}
type ComponentsConfig struct {
   ComponentArn string `json:"componentArn"`
   Parameters []struct {
      Name string `json:"name"`
      Value []string `json:"value"`
   } `json:"parameters"`}
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
         Encrypted bool `json:"encrypted"`
         Iops int `json:"iops"`
         KmsKeyId string `json:"kmsKeyId"`
         SnapshotId string `json:"snapshotId"`
         Throughput int `json:"throughput"`
         VolumeSize int `json:"volumeSize"`
         VolumeType string `json:"volumeType"`

      }
      NoDevice string `json:"noDevice"`
      VirtualName string `json:"virtualName"`
   } `json:"blockDeviceMappings"`
 
   ClientToken string `json:"clientToken"`
   Components []models.ImageComponent `json:"components"`
   Description string `json:"description"`
   Name string `json:"name"`
   ParentImage string `json:"parentImage"`
   SemanticVersion string `json:"semanticVersion"`
   Tags map[string]string `json:"tags"`
   WorkingDirectory string `json:"workingDirectory"`

}
type DeleteImageRecipeResponse struct {
   ImageRecipeArn string `json:"imageRecipeArn"`
   RequestId string `json:"requestId"`
}
 
type CreateImageComponentRequest struct {
   ChangeDescription string `json:"changeDescription"`
   ClientToken string `json:"clientToken"`
   Data string `json:"data"`
   Description string `json:"description"`
   KmsKeyId string `json:"kmsKeyId"`
   Name string `json:"name"`
   Platform string `json:"platform"`
   SemanticVersion string `json:"semanticVersion"`
   SupportedOsVersions []string `json:"supportedOsVersions"`
   Tags map[string]string `json:"tags"`
   Uri string `json:"uri"`

}
 
type CreateImageComponentResponse struct {
   ClientToken string `json:"clientToken"`
   ComponentBuildVersionArn string `json:"componentBuildVersionArn"`
   RequestId string `json:"requestId"`
}
//type DeleteImageComponentResponse struct {}
