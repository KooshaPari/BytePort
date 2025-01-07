package ec2

import (
	"context"
	"fmt"
	aws "nvms/lib/awspin"
	"time"
)


func(c *Client) GetAlbNetworkInfo(ctx context.Context )(string, string,string, string, error){
    fmt.Println("Getting VPC")
    vpcId,err := c.describeDefaultVPC(ctx)
    if err !=nil {
        return "", "", "","", err
    }
    fmt.Println("Getting Subnets")
    subnets, err := c.DescribeSubnets(ctx, vpcId)
    if err != nil {
        return "", "", "","", err
    }
    fmt.Println("Getting Security Groups")
    securityGroups, err := c.DescribeSecurityGroups(ctx, vpcId)
    if err != nil {
        return "", "", "","", err
    }
    fmt.Println("Got ALB NetInfo")
    //fmt.Println("Subnets: ", subnets)
    //fmt.Println("Security Groups: ", securityGroups)
    //fmt.Println("VPC: ", vpcId)
    subnet1, subnet2 := subnets.SubnetSet[0].SubnetId, subnets.SubnetSet[1].SubnetId;
    return subnet1, subnet2, securityGroups.SecurityGroupInfo.Item.GroupId,vpcId, nil
} 
func (c *Client) WaitForEC2Running(instanceIDs []string, ctx context.Context) error {
    maxAttempts := 60 // Adjust as needed
    fmt.Println("Waiting for EC2 to initialize: ", instanceIDs)

    for attempt := 0; attempt < maxAttempts; attempt++ {
        fmt.Printf("Attempt %d: Checking EC2 instance statuses\n", attempt+1)
        resp, err := c.DescribeInstances(ctx, instanceIDs)
        if err != nil {
            // Check if error is InvalidInstanceID.NotFound and retry
            if awsErr, ok := err.(*aws.ErrorResponse); ok && awsErr.Code == "InvalidInstanceID.NotFound" {
                fmt.Println("Instance not yet available, retrying...")
            } else {
                 fmt.Println("Instance not yet available, retrying...")
                //return fmt.Errorf("error checking instance status: %v", err)
            }
        } else {
            allRunning := true
            for _, reservation := range resp.Reservations {
                for _, instance := range reservation.Instances {
                    if instance.State.Name != "running" {
                        allRunning = false
                        break
                    }
                }
                if !allRunning {
                    break
                }
            }

            if allRunning {
                fmt.Println("EC2 instances initialized: ", instanceIDs)
                return nil
            }
        }

        // Implement a non-blocking wait without using time.Sleep
        start := time.Now()
        waitDuration := 5 * time.Second
        for {
            elapsed := time.Since(start)
            if elapsed >= waitDuration {
                break
            }
        }
    }
    fmt.Println("Timeout waiting for running instances")

    return fmt.Errorf("timeout waiting for instances to reach running state")
} 

func(c *Client) GetInfraNetworkInfo(ctx context.Context )(string, string,string,error){
    fmt.Println("Getting VPC")
    vpcId,err := c.describeDefaultVPC(ctx)
    if err !=nil {
        return "", "","",  err
    }
    fmt.Println("Getting Subnets")
    subnets, err := c.DescribeSubnets(ctx, vpcId)
    if err != nil {
        return "", "","",  err
    }
    fmt.Println("Getting Security Groups")
    securityGroups, err := c.DescribeSecurityGroups(ctx, vpcId)
    if err != nil {
        return "", "","",  err
    }
    fmt.Println("Got ALB NetInfo")
    //fmt.Println("Subnets: ", subnets)
    //fmt.Println("Security Groups: ", securityGroups)
    //fmt.Println("VPC: ", vpcId)
    subnet1, subnet2 := subnets.SubnetSet[0].SubnetId, subnets.SubnetSet[1].SubnetId;
    return subnet1,subnet2,  securityGroups.SecurityGroupInfo.Item.GroupId, nil
} 