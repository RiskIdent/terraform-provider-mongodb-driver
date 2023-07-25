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
    "SCRAM-SHA-256",
  ]
}

// With custom timeouts
resource "mongodb_user" "example" {
  user = "my-user"
  db   = "my-db"
  pwd  = "super-secret-password"

  // Timeouts default to 30 seconds
  timeouts = {
    create = "1m"    // 1 minute
    read   = "5s"    // 5 seconds
    update = "1m30s" // 1 minute & 30 seconds
    delete = "500ms" // 500 milliseconds (or 0.5 seconds)
  }
}
