create table fiat (
	id uuid varchar(16) not null,

	name varchar(64) not null,
	code varchar(8) not null
);

create table fiat_history (
	fiat_id uuid not null,

	date date not null,
	value numeric(16,16) not null,

	foreign key (fiat_id) references fiat (id)
);
