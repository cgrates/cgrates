	
db = db.getSiblingDB('admin')
db.createUser(
  {
    user: "cgrates",
    pwd: "CGRateS.org",
    roles: [ { role: "userAdminAnyDatabase", db: "admin" } ]
  }
)
