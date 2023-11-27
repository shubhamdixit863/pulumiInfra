.PHONY: all build move dockerbuild start
K8YAML_S3=s3deploy.yaml
K8YAML_API=apideploy.yaml
K8YAML_FRONTEND=frontenddeploy.yaml


AWS_REGION=us-east-2
CLUSTER_NAME=cluster-eksCluster-db268ea

all:  deployapi  configureaws

configureaws:
	  aws eks --region $(AWS_REGION)  update-kubeconfig --name  $(CLUSTER_NAME)

deployapi:
	@echo "deploying api k8 server...."
	 kubectl apply -f $(K8YAML_API)


deploys3:
	   @echo "deploying on s3 file upload k8 server...."
	   kubectl apply -f $(K8YAML_S3)

lb:
	kubectl get svc


deployfrontend:
	   @echo "deploying on s3 file upload k8 server...."
	   kubectl apply -f $(K8YAML_FRONTEND)



pulumi:
	   @echo "upping the pulumi....."
	   pulumi up