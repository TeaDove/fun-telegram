package ch_repository

var initSQL = []string{`
CREATE TABLE IF NOT EXISTS message
(
    created_at timestamp,

    tg_chat_id Int64,
    tg_id      Int64,

    tg_user_id Int64,
    text       String 
) ENGINE = ReplacingMergeTree() ORDER BY (tg_chat_id, tg_id)`,
	`ALTER TABLE message ADD COLUMN IF NOT EXISTS reply_to_msg_id Nullable(Int64)`,
	`ALTER TABLE message ADD COLUMN IF NOT EXISTS reply_to_user_id Nullable(Int64)`,
	`ALTER TABLE message ADD COLUMN IF NOT EXISTS words_count UInt64`,
	`ALTER TABLE message ADD COLUMN IF NOT EXISTS toxic_words_count UInt64`,
	`ALTER TABLE message DROP COLUMN IF EXISTS id`,
	`
CREATE TABLE IF NOT EXISTS channel
(
    tg_id 			 Int64,
    tg_title      String,
    tg_username      String,
	uploaded_at timestamp,

    participant_count 		 Int64,
    recommendations_ids      Array(Int64)
) ENGINE = ReplacingMergeTree() ORDER BY (tg_id);`, `
create table if not exists channel_edge
(
    tg_id_in  Int64,
    tg_id_out Int64,
    order     Int64

) ENGINE = ReplacingMergeTree() ORDER BY (tg_id_in, tg_id_out);`,
	`alter table channel add column if not exists is_leaf bool default false;`,
	`alter table channel add column if not exists tg_about Nullable(String);`,
}
