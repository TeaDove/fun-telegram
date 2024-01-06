db.createCollection("user")
db.event.createIndex({ "tg_chat_id": 1, "user_id": 1 }, { "unique": true })

db.createCollection("user_in_chat")
db.user_in_chat.createIndex({ "tg_user_id": 1 }, { "unique": true })
db.user_in_chat.createIndex({ "tg_username": 1 }, { "unique": true })
