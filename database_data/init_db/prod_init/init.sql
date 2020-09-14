Create table if not exists "User"
(
    user_id    bigint primary key,
    user_name  text NOT NULL,
    balance    DECIMAL(19, 4),
    created_at timestamptz
);

Create table if not exists "Operation"
(
    operation_id serial primary key,
    user_id      bigint references "User" (user_id),
    comment      text,
    amount       DECIMAL(19, 4),
    date         timestamptz
);
