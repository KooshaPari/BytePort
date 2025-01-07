package ec2
/*
HTTP/1.1 200 OK
x-amzn-RequestId: 123a4b56-7c89-01d2-3ef4-example5678f
Content-Type: application/x-amz-json-1.1
Content-Length: 339
Connection: keep-alive

{
   "repository":{
      "repositoryArn":"arn:aws:ecr:us-west-2:012345678910:repository/sample-repo",
      "registryId":"012345678910",
      "repositoryName":"sample-repo",
      "repositoryUri":"012345678910.dkr.ecr.us-west-2.amazonaws.com/sample-repo",
      "createdAt":1.563223656E9,
      "imageTagMutability":"MUTABLE",
      "imageScanningConfiguration": {
            "scanOnPush": false
      }
   }
}
*/
type CreateRepositoryResponse struct {}\
/*
*.
type DeleteRepositoryResponse struct {}
type PutImageResponse struct {}
