package ecr

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
)

 
import (
	"context"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	aws "nvms/lib/awspin"
	"sort"
	"strings"
	"time"

	spinhttp "github.com/fermyon/spin-go-sdk/http"
)

// NewEC2 creates a new EC2 Client
func NewEC2(config aws.Config) (*Client, error) {
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
func (c *Client) buildEndpoint(action string) (string, error) {
    u, err := url.Parse(c.endpointURL)
    if err != nil {
        return "", fmt.Errorf("failed to parse endpoint: %w", err)
    }

    if c.usePathStyle {
        // LocalStack: http://localhost:4566/elasticloadbalancing/
        u = u.JoinPath("elasticloadbalancing")
    }
    
    // Both AWS and LocalStack use query parameters for ELB API
    q := u.Query()
    q.Set("Action", action)
    if(q.Get("Version") == "") {
    q.Set("Version", "2015-12-01")  // ELB API version}
    }
    u.RawQuery = q.Encode()

    return u.String(), nil
}

func (c *Client) newRequest(ctx context.Context, method string, params map[string]string, body []byte) (*http.Request, error) {
    furl, err := c.buildEndpoint(params["Action"])
     if err != nil {
        return nil, err
    }
    u, err := url.Parse(furl)
    if err != nil {
        return nil, err
    }

    var awsDate aws.AwsDate
    awsDate.Time = time.Now()

    // Add required AWS Query API parameters
    if params["Version"] == "" {
    params["Version"] = "2016-11-15"}
    params["X-Amz-Algorithm"] = "AWS4-HMAC-SHA256"
    params["X-Amz-Date"] = awsDate.GetTime()
    
    // Build credential scope
    credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request",
        awsDate.GetDate(),
        c.config.Region,
        c.config.Service)
    
    params["X-Amz-Credential"] = fmt.Sprintf("%s/%s",
        c.config.AccessKeyId,
        credentialScope)

    // Set signed headers
    params["X-Amz-SignedHeaders"] = "host"

    // Add security token if present
    if c.config.SessionToken != "" {
        params["X-Amz-Security-Token"] = c.config.SessionToken
    }

    // Build canonical query string for signing
    canonicalQueryString := GetCanonicalQueryString(params)

    // Create string to sign
    canonicalRequest := strings.Join([]string{
        method,
        "/",
        canonicalQueryString,
        fmt.Sprintf("host:%s\n", u.Host),  // Canonical headers
        "host",                            // Signed headers
        "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", // Empty payload hash
    }, "\n")

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

    // Add signature to parameters
    params["X-Amz-Signature"] = signature

    // Build final URL with all parameters
    query := u.Query()
    for k, v := range params {
        query.Set(k, v)
    }
    u.RawQuery = query.Encode()

    //fmt.Printf("Request URL: %s\n", u.String())

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


// Client provides an interface for interacting with the EC2 API
type Client struct {
    config       aws.Config
    endpointURL string
    usePathStyle bool
}

// Instance represents an EC2 instance
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
func (c *Client) PutImage(ctx context.Context, name string, manifest string) (*PutImageResponse, error) {
	params := map[string]string{
		"X-Amz-Target": "AmazonEC2ContainerRegistry_V20150921.PutImage",
		"repositoryName": name,
		"imageManifest": manifest,
	}
	req, err := c.newRequest(ctx, "POST", params, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result PutImageResponse
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}
 /*POST / HTTP/1.1
Host: api.ecr.us-west-2.amazonaws.com
Accept-Encoding: identity
X-Amz-Target: AmazonEC2ContainerRegistry_V20150921.CreateRepository
Content-Type: application/x-amz-json-1.1
User-Agent: aws-cli/1.16.190 Python/3.6.1 Darwin/16.7.0 botocore/1.12.180
X-Amz-Date: 20190715T204735Z
Authorization: AUTHPARAMS
Content-Length: 33

{
   "repositoryName": "sample-repo"
}*/
func (c *Client) CreateRepository(ctx context.Context, name string) (*CreateRepositoryResponse, error) {
	requestBody := struct {
        RepositoryName string `json:"repositoryName"`
    }{
        RepositoryName: name,
    }
    
    bodyJSON, err := json.Marshal(requestBody)
    if err != nil {
        return nil, err
    }
	params := map[string]string{
		"Action": "CreateRepository",
		"Version": "2015-09-21",
		"repositoryName": name,
    }

    // Remove Version and Action - they're not used in this API style
    req, err := c.newRequest(ctx, "POST", params, bodyJSON)
    if err != nil {
        return nil, err
    }
	fmt.Println("Request created: ", req)
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
 
	var result CreateRepositoryResponse
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}
func (c *Client) DeleteRepository(ctx context.Context, name string) (*DeleteRepositoryResponse, error) {
	params := map[string]string{
		"X-Amz-Target": "AmazonEC2ContainerRegistry_V20150921.CreateRepository",
		"repositoryName": name,
		"Force": "true",
	}
	req, err := c.newRequest(ctx, "POST", params, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result DeleteRepositoryResponse
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}
/* PUT /CreateContainerRecipe HTTP/1.1
Content-type: application/json

{
   "clientToken": "string",
   "components": [ 
      { 
         "componentArn": "string",
         "parameters": [ 
            { 
               "name": "string",
               "value": [ "string" ]
            }
         ]
      }
   ],
   "containerType": "string",
   "description": "string",
   "dockerfileTemplateData": "string",
   "dockerfileTemplateUri": "string",
   "imageOsVersionOverride": "string",
   "instanceConfiguration": { 
      "blockDeviceMappings": [ 
         { 
            "deviceName": "string",
            "ebs": { 
               "deleteOnTermination": boolean,
               "encrypted": boolean,
               "iops": number,
               "kmsKeyId": "string",
               "snapshotId": "string",
               "throughput": number,
               "volumeSize": number,
               "volumeType": "string"
            },
            "noDevice": "string",
            "virtualName": "string"
         }
      ],
      "image": "string"
   },
   "kmsKeyId": "string",
   "name": "string",
   "parentImage": "string",
   "platformOverride": "string",
   "semanticVersion": "string",
   "tags": { 
      "string" : "string" 
   },
   "targetRepository": { 
      "repositoryName": "string",
      "service": "string"
   },
   "workingDirectory": "string"
}*/
/*
func (c *Client) CreateContainerRecipe(ctx context.Context) (*CreateContainerRecipeResponse, error){}
// DELETE /DeleteContainerRecipe?containerRecipeArn=containerRecipeArn HTTP/1.1
func (c *Client) DeleteContainerRecipe(ctx context.Context) (*DeleteContainerRecipeResponse, error){}*/