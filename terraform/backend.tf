terraform {
  backend "s3" {
    bucket  = "steve-rhoton-tfstate"
    key     = "steverhoton-unt-units-svc/terraform.tfstate"
    region  = "us-west-2"
    encrypt = true
  }
}