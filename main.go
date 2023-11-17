package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/rds"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/s3"
	"github.com/pulumi/pulumi-eks/sdk/go/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

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

		bucket, err := s3.NewBucket(ctx, "myBucket", nil)
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

		return nil
	})
}
