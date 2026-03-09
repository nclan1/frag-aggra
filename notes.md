# start up container
`docker compose up --build`
rebuild from scratch, not looking at cache

# enter the db container
psql -h localhost -U postgres -d fragrance_database

# list of tables
`\dt`

# info about tables
`\d <table>`

# list actual data from table
`SELECT * FROM your_table_name_here;`
