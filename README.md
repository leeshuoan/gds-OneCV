# Govtech GDS OneCV assessment

## Local Setup
### Database
1. Download [Postgres SQL](https://www.postgresql.org/download/)

2. Set up the database by running the initdb script with your own postgres `username` and `password`
```
psql -h localhost -U <username> -f initdb.sql
```
### Go Backend
1. Install [Go](https://go.dev/doc/install)

2. Set environment variables for your database credentials
```
export DB_USER=your_db_user
export DB_PASSWORD=your_db_password
export DB_NAME=your_db_name
```

3. Run the application
```
go run main.go
```
