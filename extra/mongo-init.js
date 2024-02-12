db.createCollection("users")
db.users.createIndex({ "tg_id": 1 }, { "unique": true })
db.users.createIndex({ "tg_username": 1 })
db.users.createIndex({ "created_at": 1 })

db.createCollection("messages")
db.messages.createIndex({ "tg_chat_id": 1, "tg_user_id": 1 })
db.messages.createIndex({ "tg_chat_id": 1, "tg_id": 1 }, { "unique": true })
db.messages.createIndex({ "created_at": 1 })

db.createCollection("members")
db.members.createIndex({ "tg_chat_id": 1, "tg_user_id": 1 }, { "unique": true })
