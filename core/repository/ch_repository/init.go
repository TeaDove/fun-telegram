package ch_repository

var initSQL = []string{`
CREATE TABLE IF NOT EXISTS message
(
    id         UUID NOT NULL default generateUUIDv4(),
    created_at timestamp NOT NULL,

    tg_chat_id Int64 NOT NULL,
    tg_id      Int64 NOT NULL,

    tg_user_id Int64 NOT NULL,
    text       String NOT NULL
) ENGINE = ReplacingMergeTree() ORDER BY (tg_chat_id, tg_id)`,
	`ALTER TABLE message ADD COLUMN IF NOT EXISTS reply_to_msg_id Nullable(Int64)`,
	`ALTER TABLE message ADD COLUMN IF NOT EXISTS reply_to_user_id Nullable(Int64)`,
}
