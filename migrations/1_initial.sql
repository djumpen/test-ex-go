-- +migrate Up notransaction
CREATE TYPE state AS ENUM ('WIN', 'LOSS');
CREATE TYPE status AS ENUM ('PROCESSED', 'CANCELED');

create table events
(
	id serial not null
		constraint events_pk
			primary key,
	state state not null,
	amount float not null,
	transaction_id varchar(128) not null,
	status status not null,
	created_at timestamp default now() not null,
	updated_at timestamp default now()
);

create unique index events_transaction_id_uindex
	on events (transaction_id);


create table balance
 (
 	id serial not null
 		constraint balance_pk
 			primary key,
 	total float not null,
 	updated_at timestamp default now() not null
 );

 INSERT INTO balance(total) VALUES(0);