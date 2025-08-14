CREATE TABLE outbox_messages
(
    id            bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    type          varchar(128) not null,
    payload       bytea        not null,
    created_at    timestamp(6) not null,
    trace_context bytea null
);