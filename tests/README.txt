
cd testN
terraform init
terraform plan -var-file="../test.tfvar"
terraform apply -var-file="../test.tfvar"
terraform destroy -var-file="../test.tfvar"

