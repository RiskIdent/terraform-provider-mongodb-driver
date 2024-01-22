resource "mongodb_role" "example" {
  role = "myClusterwideAdmin"
  db   = "admin"
  privileges = [
    {
      resource = { cluster = true }
      actions  = ["addShard"]
    },
    {
      resource = { db = "config", collection = "" }
      actions  = ["find", "update", "insert", "remove"]
    },
    {
      resource = { db = "users", collection = "usersCollection" },
      actions  = ["update", "insert", "remove"]
    },
    {
      resource = { db = "", collection = "" },
      actions  = ["find"]
    }
  ]
  roles = { role = "read", db = "admin" }
}
