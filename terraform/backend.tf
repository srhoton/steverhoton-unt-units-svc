terraform {
  backend "s3" {
    bucket  = "steve-rhoton-tfstate"
    key     = "sr-unit/terraform.tfstate"
    region  = "us-west-2"
    encrypt = true
  }
}