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

INSERT INTO "User" (user_id, user_name, balance, created_at)
VALUES (1, 'Mr. Smith', 0, '2020-08-11T10:23:58+03:00'),
       (2, 'Mr. Jones', 10, '2020-08-11T10:23:58+03:00');

INSERT INTO "Operation" (user_id, comment, amount, date)
VALUES (1, 'incoming payment', 10,                  '2020-08-11T10:23:58+03:00'),
       (2, 'incoming payment', 10,                  '2020-08-11T10:23:59+03:00'),
       (1, 'transfer to Mr. Jones', -10,            '2020-08-11T10:24:00+03:00'),
       (2, 'transfer from Mr. Smith', 10,           '2020-08-11T10:24:01+03:00'),
       (2, 'payment to advertisement service', -10, '2020-08-11T10:24:02+03:00')