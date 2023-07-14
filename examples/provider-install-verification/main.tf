
terraform {
  required_providers {
    mongodb = {
      source = "hashicorp.com/edu/mongodb-driver"
    }
  }
}

provider "mongodb" {
  host = "http://localhost:19090"
  username = "education"
  password = "test123"
}

data "mongodb_coffees" "example" {}
