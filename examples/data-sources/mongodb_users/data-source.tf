// Search all users in all databases
data "mongodb_users" "example" {
}

// Search all users in specific databases
data "mongodb_users" "example" {
  db = "my-db"
}

// Search all users with custom filter
data "mongodb_users" "example" {
  filter = {
    "customData.my-custom-field" = "my-custom-value"
  }
}

// With custom timeouts
data "mongodb_users" "example" {
  db = "my-db"

  // Timeouts default to 30 seconds
  timeouts = {
    read = "5s" // 5 seconds
  }
}
