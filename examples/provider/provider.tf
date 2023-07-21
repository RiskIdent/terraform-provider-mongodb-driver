provider "mongodb" {
  uri = "mongodb://localhost:27017"
}

// With username & password
provider "mongodb" {
  uri      = "mongodb://my-user:my-password@localhost:27017"
  username = "my-user"
  password = "my-password"
}
