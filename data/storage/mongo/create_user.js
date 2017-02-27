db = db.getSiblingDB('cgrates')
db.createUser(
  {
    user: "cgrates",
    pwd: "CGRateS.org",
    roles: [ { role: "dbAdmin", db: "cgrates" } ]
  }
)

