all: clean
	go build

clean:
	rm -rf database.db* backup?.db*
