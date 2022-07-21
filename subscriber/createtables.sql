drop table if exists payment;
drop table if exists delivery;
drop table if exists items;
drop table if exists orders;

create table orders (
Id                  serial primary key not null,
order_uid           varchar(45) not null unique,
track_number        varchar(45),
entry               varchar(45),
locale              varchar(45),
internal_signature  varchar(45),
customer_id         varchar(45),
delivery_service    varchar(45),
shardkey            varchar(45),
sm_id               int,
date_created        timestamp with time zone,
oof_shard           varchar(45)
);


create table items ( 
id              serial primary key not null,
orderid         varchar(45) references orders(order_uid) on delete cascade,
chrt_id         int,
track_number    varchar(45),
price           numeric,
rid             varchar(45),
name            varchar(45),
sale            int,
item_size       varchar(45),
total_price     numeric,
nm_id           numeric,
brand           varchar(45),
status          int
);

create table payment ( 
id                  serial primary key not null,
orderid             varchar(45) references orders(order_uid) on delete cascade,
transaction         varchar(45),
request_id          varchar(45),
currency            varchar(5),
provider            varchar(45),
amount              int,
payment_dt          numeric,
bank                varchar(45),
delivery_cost       int,
goods_total         int,
custom_fee          int
);

create table delivery ( 
id              serial primary key not null,
orderid         varchar(45) references orders(order_uid) on delete cascade,
name            varchar(45),
phone           varchar(45),
zip             varchar(45),
city            varchar(45),
address         varchar(45),
region          varchar(45),
email           varchar(45)
);

