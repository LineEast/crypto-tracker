create table fiat (
	id uuid default gen_random_uuid() primary key,

	fiat_id varchar(16) not null,
	name varchar(64) not null,
	code varchar(8) not null
);

create table fiat_history (
	fiat_id uuid not null,

	date date not null,
	value numeric(16,16) not null,

	foreign key (fiat_id) references fiat (id)
);

create table btc_usdt_history (
	btc_usdt_id uuid not null,

	fiat_id uuid not null,
	value numeric(16, 16),

	foreign key (fiat_id) references fiat (id),
	foreign key (btc_usdt_id) references btc_usdt (id)
);

create table btc_usdt (
	id uuid default gen_random_uuid() primary key,

	date integer not null,
	avarage_price numeric(16,16) not null
);
