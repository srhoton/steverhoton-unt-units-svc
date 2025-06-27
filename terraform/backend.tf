terraform {
  backend "s3" {
    bucket  = "srhoton-tfstate"
    key     = "steverhoton-unt-units-svc/terraform.tfstate"
    region  = "us-east-1"
    encrypt = true
  }
}