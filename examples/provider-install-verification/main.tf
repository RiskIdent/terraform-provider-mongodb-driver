
terraform {
  required_providers {
    mongodb = {
      source = "hashicorp.com/edu/mongodb-driver"
    }
  }
}

provider "mongodb" {}

data "mongodb_coffees" "example" {}
