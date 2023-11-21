package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/rds"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/s3"
	"github.com/pulumi/pulumi-eks/sdk/go/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"log"
)

const BUCKET_NAME = "gorestapi"

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		desiredCapacity := pulumi.Int(2)
		minSize := pulumi.Int(2)
		maxSize := pulumi.Int(2)
		instanceType := pulumi.String("t2.medium")
		// Create an EKS cluster with the default configuration.
		cluster, err := eks.NewCluster(ctx, "cluster", &eks.ClusterArgs{
			DesiredCapacity: desiredCapacity,
			MinSize:         minSize,
			MaxSize:         maxSize,
			InstanceType:    instanceType,
		})
		if err != nil {
			return err
		}
		// Export the cluster's kubeconfig.
		ctx.Export("kubeconfig", cluster.Kubeconfig)

		bucket, err := s3.NewBucket(ctx, BUCKET_NAME, nil)
		if err != nil {
			return err
		}
		_, err = s3.NewBucketPublicAccessBlock(ctx, "bucketPublicAccessBlock", &s3.BucketPublicAccessBlockArgs{
			Bucket: bucket.ID(),

			BlockPublicAcls:       pulumi.Bool(true),
			BlockPublicPolicy:     pulumi.Bool(false),
			IgnorePublicAcls:      pulumi.Bool(true),
			RestrictPublicBuckets: pulumi.Bool(false),
		})
		if err != nil {
			return err
		}

		bucketPolicy := iam.GetPolicyDocumentOutput(ctx, iam.GetPolicyDocumentOutputArgs{
			Statements: iam.GetPolicyDocumentStatementArray{
				&iam.GetPolicyDocumentStatementArgs{
					Actions: pulumi.StringArray{
						pulumi.String("s3:GetObject"),
					}, Conditions: nil,
					Effect: pulumi.String("Allow"),
					Principals: iam.GetPolicyDocumentStatementPrincipalArray{
						&iam.GetPolicyDocumentStatementPrincipalArgs{
							Type: pulumi.String("AWS"),
							Identifiers: pulumi.StringArray{
								pulumi.String("*"),
							},
						},
					},
					Resources: pulumi.StringArray{
						pulumi.Sprintf("%s/*", bucket.Arn),
					},
					Sid: pulumi.String("PublicRead"),
				},
			},
		})

		// Attach the policy to the bucket
		_, err = s3.NewBucketPolicy(ctx, "bucketPolicy", &s3.BucketPolicyArgs{
			Bucket: bucket.ID(),
			Policy: bucketPolicy.ApplyT(func(allowAccessFromAnotherAccountPolicyDocument iam.GetPolicyDocumentResult) (*string, error) {
				return &allowAccessFromAnotherAccountPolicyDocument.Json, nil
			}).(pulumi.StringPtrOutput),
		})
		if err != nil {
			return err
		}

		ctx.Export("bucketName", bucket.ID())

		// Create a new security group for RDS
		secGroup, err := ec2.NewSecurityGroup(ctx, "rdsSecGroup", &ec2.SecurityGroupArgs{
			Ingress: ec2.SecurityGroupIngressArray{
				ec2.SecurityGroupIngressArgs{
					Protocol: pulumi.String("tcp"),
					FromPort: pulumi.Int(5432),
					ToPort:   pulumi.Int(5432),
					CidrBlocks: pulumi.StringArray{
						pulumi.String("0.0.0.0/0"),
					},
				},
			},
		})
		if err != nil {
			return err
		}

		// Create RDS instance
		rdsInstance, err := rds.NewInstance(ctx, "goapirds", &rds.InstanceArgs{
			InstanceClass:      pulumi.String("db.t2.micro"),
			AllocatedStorage:   pulumi.Int(20),
			Engine:             pulumi.String("postgres"),
			EngineVersion:      pulumi.String("12"),
			Username:           pulumi.String("myusername"),
			Password:           pulumi.String("mypassword"),
			ParameterGroupName: pulumi.String("default.postgres12"),
			VpcSecurityGroupIds: pulumi.StringArray{
				secGroup.ID(),
			},
			PubliclyAccessible: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}
		ctx.Export("rdsEndpoint", rdsInstance.Endpoint)
		log.Println("RDS Endpoint", rdsInstance.Endpoint)

		return nil
	})
}
