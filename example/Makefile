main: example brief.db
	./example brief.db

example: main.go models.go
	go build

brief.db: brief.schema.sql brief.data.sql
	cat brief.schema.sql brief.data.sql | sqlite3 brief.db

clean:
	rm example brief.db
