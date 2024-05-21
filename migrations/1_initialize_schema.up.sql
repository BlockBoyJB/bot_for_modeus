create table if not exists "user" (
    id int generated always as identity primary key,
    user_id int unique not null,
    full_name varchar not null,
    login varchar,
    password varchar
)