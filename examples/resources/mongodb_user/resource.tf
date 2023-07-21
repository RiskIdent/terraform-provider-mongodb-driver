resource "mongodb_user" "example" {
  user = "my-user"
  db   = "my-db"
  pwd  = "super-secret-password"
}

// With role
resource "mongodb_user" "example" {
  user = "my-user"
  db   = "my-db"
  pwd  = "super-secret-password"

  roles = [
    // Only specifying "role" will use the same database as the user
    {
      role = "readWrite"
    },
    // Reference role in other database
    {
      role = "readWrite"
      db   = "admin"
    },
  ]
}

// With customData
resource "mongodb_user" "example" {
  user = "my-user"
  db   = "my-db"
  pwd  = "super-secret-password"

  customData = {
    "my-custom-field" = "my-custom-value"
  }
}

// With explicit mechanisms
resource "mongodb_user" "example" {
  user = "my-user"
  db   = "my-db"
  pwd  = "super-secret-password"

  mechanisms = [
    "SCRAM-SHA-1",
    "SCRAM-SHA-256",
  ]
}
