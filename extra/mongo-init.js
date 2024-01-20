db.createCollection("users")
db.users.createIndex({ "tg_user_id": 1 }, { "unique": true })
db.users.createIndex({ "tg_username": 1 }, { "unique": false })

db.createCollection("messages")
db.messages.createIndex({ "tg_chat_id": 1, "tg_user_id": 1 }, { "unique": false })
