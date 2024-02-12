package ch_repository

var initSQL = []string{`
CREATE TABLE IF NOT EXISTS message
(
    id         UUID not null default generateUUIDv4(),
    created_at timestamp not null,

    tg_chat_id Int64 not null,
    tg_id      Int64 NOT NULL,

    tg_user_id Int64 not null,
    text       String NOT NULL
) ENGINE = ReplacingMergeTree() primary key (tg_chat_id, tg_id)
`}
