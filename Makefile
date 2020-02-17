build:
	go build -o tresor main.go

clean:
	rm -rf tresor

install: build
	cp tresor ~/.local/bin/tresor
	chmod +x ~/.local/bin/tresor
