
create table application (
  id serial primary key,
  version integer not null,
  application_name varchar(100) not null
);

insert into application (version, application_name) values (0, 'localiday');

create table users (
  id serial primary key,
  username varchar(150),
  password varchar(150),
  full_name varchar(250),
  nickname varchar(75),
  email varchar(200),
  password_expired boolean,
  active boolean,
  constraint users_username_unq unique(username),
  constraint users_email_unq unique(email)
);

create unique index users_username_password_idx on users(username, password);
create index users_active_idx on users(active);

create table roles (
  id serial primary key,
  authority varchar(100) not null,
  constraint roles_authority_idx unique(authority)
);

create table user_roles (
  id serial primary key,
  user_id integer references users(id) not null,
  role_id integer references roles(id) not null
);

create unique index user_roles_map_idx on user_roles(user_id, role_id);

update application set version = 1 where application_name = 'localiday';