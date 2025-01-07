package lib

import (
	"context"

	"fmt"
	aws "nvms/lib/awspin"
	ec2 "nvms/lib/awspin/ec2"
	ecr "nvms/lib/awspin/ec2/ecr"
	img "nvms/lib/awspin/ec2/imaging"
	awsnet "nvms/lib/awspin/network"
	r53 "nvms/lib/awspin/network/route53"
	"nvms/lib/awspin/s3"
	"nvms/models"
	"strings"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)


var AWSEndpointBase string = "https://%s.us-east-1.amazonaws.com" /* "http://localhost.localstack.cloud:4566"*/
func PushToS3(zipBall []byte, AccessKey string, SecretKey string, ProjectName string) (S3DeploymentInfo,error) {
	fmt.Println("Uploading to S3...")
 	cfg := aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: getServiceEndpoint("s3"),
		Region: "us-east-1",
		Service: "s3",
	}
	ctx := context.Background()
	s3Client, err := s3.NewS3(cfg)
	if err != nil {
		fmt.Println(err)
		return S3DeploymentInfo{},err
	}
	fmt.Println("Created S3 client")
	bucketName := strings.ToLower(ProjectName) + "-bytebucket-" + uuid.New().String()
	err = s3Client.CreateBucket(ctx, bucketName)
	if err != nil {
			fmt.Println(err)
			return S3DeploymentInfo{},err
		}
		
	fmt.Println("Created bucket")
	err = s3Client.PutObject(ctx, bucketName, "src.zip", zipBall)
	if err != nil {
		fmt.Println(err)
		return S3DeploymentInfo{},err
	}
	
	// return uri/bucket name for later use
	
	info := S3DeploymentInfo{
        BucketName:  bucketName,
        ObjectKey:   "src.zip",
        Region:      "us-east-1",
        BucketARN:   fmt.Sprintf("arn:aws:s3:::%s", bucketName),
		//ObjectURL:   fmt.Sprintf("http://localhost:4566/%s/%s", bucketName, "src.zip"),
		ObjectURL:   fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, "src.zip"),
        ContentHash: aws.GetPayloadHash(zipBall),
    }
	fmt.Println("Uploaded to S3")
	return info,nil
	
}


func DeployEC2(AccessKey string, SecretKey string, bucket S3DeploymentInfo, service models.Service, fileMap []string) ([]EC2InstanceInfo,error){
	client, err := ec2.NewEC2(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: getServiceEndpoint("ec2"),
		Region: "us-east-1",
		Service: "ec2",
	})
	if err != nil {
		fmt.Println(err)
		return []EC2InstanceInfo{},err
	}
	
	buildScript,err := generateBuildScript(bucket, service,AccessKey, SecretKey, fileMap)
	if err != nil {
		fmt.Println("Error generating build script: ", err)
		return []EC2InstanceInfo{}, err
	}
	//fmt.Println("Generated build script: ", buildScript)
	fmt.Println("Created EC2 client: ", client)
	params := map[string]string{
		"ImageId": "ami-01816d07b1128cd2d",
		//"ImageId": "ami-024f768332f0",
		"InstanceType": "t2.micro",
		"UserData": buildScript,
		"MinCount": "1",
		"MaxCount": "1",
	}
	fmt.Println("Creating instance")
	resp, err := client.RunInstances(context.Background(), params)
	//fmt.Println(resp)
	var instances []EC2InstanceInfo
    for _, instance := range resp.Instances {
        newInstance := EC2InstanceInfo{
            InstanceID:     instance.InstanceId,
            PrivateIP:      instance.PrivateIpAddress,
            State:          instance.State.Name,
            Region:        "us-east-1",
        }
        instances = append(instances, newInstance)
    }
	return instances,nil;
}




func ProvisionNetwork(AccessKey string, SecretKey string, projectName string ) ( *awsnet.CreateLoadBalancerResponse, string, string, error) {
	albClient, err := awsnet.NewALB(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: getServiceEndpoint("elasticloadbalancing"),
		Region: "us-east-1",
		Service: "elasticloadbalancing",
	})
	if err != nil {
		fmt.Println(err)
		return nil,"","",  err
	}
	ec2Client, err := ec2.NewEC2(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: getServiceEndpoint("ec2"),
		Region: "us-east-1",
		Service: "ec2",})
		if err != nil {
			fmt.Println(err)
			return nil,"","", err
		}
	 subnet1, subnet2,sgId, vpcId, err := ec2Client.GetAlbNetworkInfo(context.Background() )
	if err != nil {
		fmt.Println(err)
		return nil,"","", err
	}
	/*targetArn, err := albClient.CreateTargetGroup(context.Background(), base+"-"+projectName+"-Byteport", vpcId)
	if err != nil {
		fmt.Println(err)
		return  "","",err
	}*/
	//fmt.Println("VPC: ", vpcId);
	albInstance, err := albClient.CreateInternetApplicationLoadbalancer(context.Background(), projectName, sgId, subnet1, subnet2)
	if err != nil {
		fmt.Println(err)
		return nil,"","", err
	}
	 //loadBalancerArn := albInstance.CreateLoadBalancerResult.LoadBalancers.Member.LoadBalancerArn
	  publicDNS := albInstance.CreateLoadBalancerResult.LoadBalancers.Member.DNSName
	 
	// for each service create targetgroup service-TG -> ALB Listener Rule Path /service/* -> service-TG
	
	fmt.Println("Hosted zone created successfully.: ", publicDNS)
	
	 
	return albInstance, vpcId, publicDNS, nil
	
}
func CreateALBListener(AccessKey string, SecretKey string, projectName string, loadBalancerArn string, vpcId string, instanceId string, port int  ) (string, string,error ){
	albClient, err := awsnet.NewALB(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: getServiceEndpoint("elasticloadbalancing"),
		Region: "us-east-1",
		Service: "elasticloadbalancing",
	})
	targetArn, err := RegisterService(AccessKey, SecretKey, loadBalancerArn, projectName, "main", vpcId, instanceId, port)
	listenerResponse,err := albClient.CreateListener(context.Background(), projectName, loadBalancerArn, targetArn)
	 if err != nil {
		fmt.Println(err)
		return "","",err
	}
	listenerArn := listenerResponse.CreateListenerResult.Listeners.Member.ListenerArn
	fmt.Println("Listener created successfully:  ", listenerArn)
	return listenerArn,targetArn,nil

}
func SetListenerRules(AccessKey string, SecretKey string, ListenerArn string, TargetArn string, serviceName string, priority int) (error){
	c, err := awsnet.NewALB(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: getServiceEndpoint("elasticloadbalancing"),
		Region: "us-east-1",
		Service: "elasticloadbalancing",
	})
	if err != nil {
		fmt.Println(err)
		return err}
	err = c.CreateListenerRule(context.Background(), ListenerArn, TargetArn, serviceName, priority)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
func RegisterService(AccessKey string, SecretKey string, loadBalancerArn string, projectName string, serviceName string, vpcId string, instanceId string, port int) (string,error){
	albClient, err := awsnet.NewALB(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: getServiceEndpoint("elasticloadbalancing"),
		Region: "us-east-1",
		Service: "elasticloadbalancing",
	})
	if err != nil {
		fmt.Println(err)
		return  "",err
	}
	targetArn, err := albClient.CreateTargetGroup(context.Background(), serviceName+"-"+projectName+"-Byteport", vpcId)
	if err != nil {
		fmt.Println(err)
		return  "",err
	}
	err =albClient.RegisterTarget(context.Background(), targetArn, instanceId, port )
	if err != nil {
		fmt.Println(err)
		return  "",err
	}
	

	return targetArn, nil;
}
func AddNewRecord(AccessKey string, SecretKey string, domainName string, zoneID string, projectName string, value string) (string, error) {
	c, err := r53.NewRoute53(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: getServiceEndpoint("route53"),
		Region: "us-east-1",
		Service: "route53",
	})
	if err != nil {
		fmt.Println(err)
		return "",err
	}

	err = c.CreateRecordSet(context.Background(), zoneID, domainName, "A", value, 300, projectName)
	if err != nil {
		fmt.Println(err)
		return "",err
	}
	fmt.Println("Record set created successfully.")
	return "Success",nil
	
}
func AwaitInitialization(AccessKey string, SecretKey string, instanceIDs []string) (error){
	fmt.Println("Waiting for instances to initialize...")
	c, err := ec2.NewEC2(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: getServiceEndpoint("ec2"),
		Region: "us-east-1",
		Service: "ec2"})
	if err != nil {
		fmt.Println(err)
		return err }
		fmt.Println("Created EC2 client")
	err = c.WaitForEC2Running(instanceIDs, context.Background())
	if err != nil {
		fmt.Println(err)
		return err}
	fmt.Println("Instances initialized")
	return nil
}
func AwaitIAMInitialization(AccessKey string, SecretKey string, instProfName string) (error){
	fmt.Println("Waiting for instances to initialize...")
	c, err := ec2.NewEC2(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: "https://iam.amazonaws.com",
		Region: "us-east-1",
		Service: "iam"})
	if err != nil {
		fmt.Println(err)
		return err }
		fmt.Println("Created IAM client")
	err = c.WaitForIAMRunning(instProfName, context.Background())
	if err != nil {
		fmt.Println(err)
		return err}
	fmt.Println("Instance Profile initialized")
	return nil
}
func TerminateS3(resource models.AWSResource, AccessKey string, SecretKey string)(error){
	c, err := s3.NewS3(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: getServiceEndpoint("s3"),
		Region: "us-east-1",
		Service: "s3",
	})
	if err != nil {
		fmt.Println(err)
		return  err
	}
	
	err = c.DeleteBucket(context.Background(), resource.ID)
	if err != nil {
		err = c.DeleteObject(context.Background(), resource.ID, "src.zip")
		if err != nil {
			fmt.Println(err)
			return err
		}
		err = c.DeleteBucket(context.Background(), resource.ID)
		if err != nil {
		fmt.Println(err)
		return err}
	}
	fmt.Println("Record set created successfully.")
	return nil
}  
func TerminateEC2(resource models.AWSResource, AccessKey string, SecretKey string)(error){
	c, err := ec2.NewEC2(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: getServiceEndpoint("ec2"),
		Region: "us-east-1",
		Service: "ec2",
	})
	if err != nil {
		fmt.Println(err)
		return  err
	}

	err = c.TerminateInstances(context.Background(), []string{resource.ID})
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Record set created successfully.")
	 
	return nil
}
func TerminateALB(resource models.AWSResource, AccessKey string, SecretKey string)(error){
	c, err := awsnet.NewALB(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: getServiceEndpoint("elasticloadbalancing"),
		Region: "us-east-1",
		Service: "elasticloadbalancing",
	})
	if err != nil {
		fmt.Println(err)
		return  err
	}

	err = c.DeleteLoadbalancer(context.Background(), resource.ID)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Record set created successfully.")
	 return nil
}
func TerminateTargetGroup(resource models.AWSResource, AccessKey string, SecretKey string)(error){
	c, err := awsnet.NewALB(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: getServiceEndpoint("elasticloadbalancing"),
		Region: "us-east-1",
		Service: "elasticloadbalancing",
	})
	if err != nil {
		fmt.Println(err)
		return  err
	}

	err = c.DeleteTargetGroup(context.Background(), resource.ID)
	if err != nil {
		fmt.Println(err)
		return err
	}
	
	fmt.Println("Record set created successfully.")
	return nil
}
func CreateECRRepo(AccessKey string, SecretKey string, projectName string) (*ecr.CreateRepositoryResponse, error) {
	fmt.Println("Creating ECR repository...")
	c, err := ecr.NewECR(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint:   "https://api.ecr.us-east-1.amazonaws.com",
		Region: "us-east-1",
		Service: "ecr",
	})
	if err != nil {
		fmt.Println(err)
		return nil,err}
   repoName :=  projectName+"-byteport-" + uuid.New().String()
   repo, err := c.CreateRepository(context.Background(), repoName)
   if err != nil {
	   fmt.Println(err)
	   return nil,err
   }
  
   fmt.Println("Repository created successfully.")
   return repo ,nil
}
func CreateImgPipeline(AccessKey string, SecretKey string, projectName string, infraARN string, recipeArn string) (*img.CreatePipelineResponse, error) {
	c, err := img.NewIMG(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint:   "https://codepipeline.us-east-1.amazonaws.com",
		Region: "us-east-1",
		Service: "codepipeline",
		
	})
	if err != nil {
		fmt.Println(err)
		return nil,err
   }
   pipelineName := projectName+"-byteport-" + uuid.New().String()
   pipeline, err := c.CreatePipeline(context.Background(), pipelineName, infraARN, recipeArn)
   if err != nil {
	   fmt.Println(err)
	   return nil,err
   }
   fmt.Println("Pipeline created successfully.")
   return  pipeline,nil
}
func CreateInfrastructureConfiguration(AccessKey string, SecretKey string, projectName string, instProf string, sgIds []string, subnetId string) (*img.CreateInfrastructureConfigurationResponse, error) {
	fmt.Println("Profile: '", instProf,"'")
	c, err := img.NewIMG(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint:   "https://imagebuilder.us-east-1.amazonaws.com",
		Region: "us-east-1",
		Service: "imagebuilder",
	})
	if err != nil {
		fmt.Println(err)
		return nil,err
   }
   infraName := projectName+"-byteport-" + uuid.New().String()
   infra, err := c.CreateInfrastructureConfiguration(context.Background(), infraName, instProf, sgIds, subnetId)
   if err != nil {
	   fmt.Println(err)
	   return nil,err
   }
   fmt.Println("Infrastructure created successfully.")
   return  infra,nil  
}
 
func ExecuteImgPipeline(AccessKey string, SecretKey string, pipelineArn string )(img.ExecuteImgPipelineResponse, error){
	c, err := img.NewIMG(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint:   "https://codepipeline.us-east-1.amazonaws.com",
		Region: "us-east-1",
		Service: "codepipeline",
	})
	if err != nil {
		fmt.Println(err)
		return img.ExecuteImgPipelineResponse{},err
   }
 
   pipeline, err := c.ExecuteImgPipeline(context.Background(), pipelineArn)
   if err != nil {
	   fmt.Println(err)
	   return img.ExecuteImgPipelineResponse{},err
   }
   fmt.Println("Pipeline created successfully.")
   return  *pipeline,nil
}
func CreateInstanceProfile(AccessKey string, SecretKey string, projectName string )(*ec2.CreateInstanceProfileResponse, error){
	c, err := ec2.NewEC2(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint:   "https://iam.amazonaws.com",
		Region: "us-east-1",
		Service: "iam",
	})
	if err != nil {
		fmt.Println(err)
		return nil,err
   }
   instProfName := projectName+"-byteport-" + uuid.New().String()
   instProf, err := c.CreateInstanceProfile(context.Background(), instProfName)
   if err != nil {
	   fmt.Println(err)
	   return nil, err
   }
   fmt.Println("Instance Profile created successfully.")
   return  instProf,nil  

}
func CreateContainerRecipe(AccessKey string, SecretKey string, workdir string, name string, components []img.ComponentsConfig ,targetRepo img.ContainerRepo) (*img.CreateContainerRecipeResponse, error) {
	c, err := img.NewIMG(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint:   "https://imagebuilder.us-east-1.amazonaws.com",
		Region: "us-east-1",
		Service: "imagebuilder",
	})
	if err != nil {
		fmt.Println(err)
		return nil,err
   }
   recipeName := name+"-byteport-" + uuid.New().String()
   recipe, err := c.CreateContainerRecipe(context.Background(), components,recipeName,  targetRepo, workdir)
   if err != nil {
	   fmt.Println(err)
	   return nil,err
   }
   fmt.Println("Container Recipe created successfully.")
   return  &recipe, nil
}
func CreateImageRecipe(AccessKey string, SecretKey string, workdir string, name string, components []img.ComponentsConfig ,targetRepo img.ContainerRepo) (*img.CreateContainerRecipeResponse, error) {
	c, err := img.NewIMG(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint:   "https://imagebuilder.us-east-1.amazonaws.com",
		Region: "us-east-1",
		Service: "imagebuilder",
	})
	if err != nil {
		fmt.Println(err)
		return nil,err
   }
   recipeName := name+"-byteport-" + uuid.New().String()
   recipe, err := c.CreateO,ageRecipe(context.Background(), components,recipeName,  targetRepo, workdir)
   if err != nil {
	   fmt.Println(err)
	   return nil,err
   }
   fmt.Println("Image Recipe created successfully.")
   return  &recipe, nil
}
 type testYaml struct{
	 Name string `yaml:"name"`
	 SchemaVersion float32 `yaml:"schemaVersion"`
	 Description string `yaml:"description"`
 }
func CreateImageComponents(AccessKey string, SecretKey string, workdir string, name string,s3Info S3DeploymentInfo, service models.Service, files []string) ([]img.ComponentsConfig, error) {
	fmt.Println("Creating Image Components...")
	c, err := img.NewIMG(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint:   "https://imagebuilder.us-east-1.amazonaws.com",
		Region: "us-east-1",
		Service: "imagebuilder",
	})
	if err != nil {
		fmt.Println(err)
		return []img.ComponentsConfig{},err
   }
   fmt.Println("Generating Image Components data...")
   imgData,err := GenerateImageComponentData(s3Info,service, AccessKey,SecretKey,files)
   if err != nil {
	   fmt.Println(err)
	   return []img.ComponentsConfig{},err
   }
   if imgData == nil {
	   fmt.Println("Image component data is nil")
	   return []img.ComponentsConfig{}, fmt.Errorf("image component data is nil")
   }
   fmt.Printf("Generated Image Components data:\n%+v\n", imgData)
   // print as YAML format
   // out put test data
   
	  var data []byte
    data, err = yaml.Marshal(imgData)
	if err != nil {
		fmt.Println("Error Marshalling YAML: ",err)
		return []img.ComponentsConfig{},err
	}
	fmt.Printf("Marshalled: \n %v\n", string(data))
 
   components := []img.ComponentsConfig{}
   for _, comp := range []string{"base", "dependencies", "app"} {
	fmt.Println("Creating Image Component: ", comp)
	   compName := name+"-"+comp
	   comp, err := c.CreateImageComponent(context.Background(), compName,"1.0.0", "Linux", string(data))
	   if err != nil {
		   fmt.Println(err)
		   return []img.ComponentsConfig{},err
	   }
	     newComp := img.ComponentsConfig{
			ComponentArn: comp.ComponentBuildVersionArn,
		 }
	   components = append(components, newComp)
   }
   fmt.Println("Image Components created successfully.")
   return  components,nil
}
func PushImagetoECR(AccessKey string, SecretKey string, image string, repo string) (*ecr.PutImageResponse, error){
	c, err := ecr.NewECR(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint:   "https://api.ecr.us-east-1.amazonaws.com",
		Region: "us-east-1",
		Service: "ecr",
	})
	if err != nil {
		fmt.Println(err)
		return nil,err
   }
   resp, err := c.PutImage(context.Background(), image, repo)
   if err != nil {
	   fmt.Println(err)
	   return nil,err
   }
   fmt.Println("Image pushed to ECR successfully.")
   return  resp,nil
}
func PrepareProject(accesskey string, secretkey string, name string, codebase []byte) (*S3DeploymentInfo,*ec2.CreateInstanceProfileResponse,*ecr.CreateRepositoryResponse ,string,string,string, error) {
	bucket, err := PushToS3(codebase, accesskey, secretkey, name)
	 
	if err != nil {
		fmt.Println("Error pushing to S3: ", err)
	 
		return nil, nil,nil,"","","",err
	}
	ecrRepo,err := CreateECRRepo(accesskey, secretkey, name)
		if err != nil {
			fmt.Println("Error creating ECR Repo: ", err)
			 
	 
			return nil, nil,nil,"","","", err
		}
	instanceProfile, err := CreateInstanceProfile(accesskey, secretkey, name)
		if err != nil {
			fmt.Println("Error creating instance profile: ", err)
 
			return nil, nil,nil, "","","",err}
 
	_, subnet1,sgId, accessURL,  err := ProvisionNewNetwork(accesskey, secretkey, name)
	if err != nil {
		fmt.Println(err)
		return nil, nil,nil,"","","", err
	}
	return &bucket,instanceProfile,ecrRepo, sgId,subnet1,accessURL,  nil
	
}
func ProvisionNewNetwork(AccessKey string, SecretKey string, projectName string ) ( *awsnet.CreateLoadBalancerResponse, string, string,string, error) {
	albClient, err := awsnet.NewALB(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: getServiceEndpoint("elasticloadbalancing"),
		Region: "us-east-1",
		Service: "elasticloadbalancing",
	})
	if err != nil {
		fmt.Println(err)
		return nil,"","","",  err
	}
	ec2Client, err := ec2.NewEC2(aws.Config{
		AccessKeyId: AccessKey,
		SecretAccessKey: SecretKey,
		SessionToken: "",
		Endpoint: getServiceEndpoint("ec2"),
		Region: "us-east-1",
		Service: "ec2",})
		if err != nil {
			fmt.Println(err)
			return nil,"","","", err
		}
	subnet1, subnet2,sgId, _, err := ec2Client.GetAlbNetworkInfo(context.Background() )
	if err != nil {
		fmt.Println(err)
		return nil,"","","", err
	}
 
	albInstance, err := albClient.CreateInternetApplicationLoadbalancer(context.Background(), projectName, sgId, subnet1, subnet2)
	if err != nil {
		fmt.Println(err)
		return nil,"","","", err
	}
 
	  publicDNS := albInstance.CreateLoadBalancerResult.LoadBalancers.Member.DNSName
 
	
	fmt.Println("Hosted zone created successfully.: ", publicDNS)
	
	 
	return albInstance, subnet1,sgId,  publicDNS, nil
	
}