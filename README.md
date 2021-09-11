# Balances service

REST API server with PostgreSQL database for RUB account balances handling. 
All operations in cents (russian kopeks).

## API description

[Postman Documentation](https://documenter.getpostman.com/view/17185878/U16kq5Hu)

| Method  | Route | JSON request | JSON response | What |
| --- | --- | --- | --- | --- |
| GET | `/` | - | - | echo |
| GET | `/balance` | `{"id": 1 }` | `{"cents_sum": 1742}` | account balance by account id |
| GET | `/balance?currency=${CURRENCY}` |  `{"id": 1 }` | `{"cents_sum": 1742}` | account balance by account id, where `CURRENCY` from https://exchangeratesapi.io/ supported currencies |
| POST | `/add` | `{"id": 1, "cents_sum": 1742}`| - | add sum of 17.42 RUB |
| POST | `/withdraw` | `{"id": 1, "cents_sum": 314}`| - | withdraw sum of 3.14 RUB |
| POST | `/transfer` | `{"sender_id": 1, "recipient_id": 2, "cents_sum": 2.17}`| - | transfer sum of 2.14 RUB from account 1 to 2 |

First `/add` creates account with such id. 

# Build and deploy

### Server and database

```bash
$ sudo docker-compose -f docker-compose.yml up
```

Uploads and starts two docker images: [server](https://hub.docker.com/repository/docker/vnch/balances_server)
and [PostgreSQL](https://hub.docker.com/repository/docker/vnch/balances_db) database.

Server daemon waits PostgreSQL initialization finish.

```bash
# docker-compose log output after up command
balances_pg_db       | 2021-08-30 10:00:46.295 UTC [1] LOG:  database system is ready to accept connections
balances_server | time="10:00:47 2021-09-10" level=info msg="trying to connect to database #1"
balances_server | time="10:00:47 2021-09-10" level=info msg="database connection established"
balances_server | time="10:00:47 2021-09-10" level=info msg="daemon started"
balances_server | time="10:00:47 2021-09-10" level=info msg="listening :9876"
```

After successful deployment you can request service by default port `9876`.

Server and database images setup available in `Dockerfile_server` and `Dockerfile_db` files.

You can build and run server locally without docker with your PostgreSQL database with schema from `api/db_schema.sql`.
```bash
$ ./build_server.sh
```

## Configuration

### Server and database

Server has config file (`configs/config.yml`) that mounts to server image in `docker-compose.yml`.

```yml
# configs/config.yml

server:
  host:            # default host
  port: 9876       # default port


bank:
  add_tries_count: 2       # SQL transaction tries count
  withdraw_tries_count: 2
  transfer_tries_count: 2

  rates_api_token:         # API token from https://exchangeratesapi.io/ for /balance?currency=EUR requests

database:
  host: database   # default db host in docker-compose
  port: 5432       # default db port in docker-compose

  name: docker     # \
  user: docker     #   db name and user info from db image `Dockerfile_db`
  password: docker # /

  conn_try_time: 5    # server db connection try duration  
  conn_tries_cnt: 10  # server db connection tries count  

  max_open_conns: 16  # golang database/sql driver
  max_idle_conns: 16  # golang database/sql driver

```

Database stores data in mounted directory `db/data`.
