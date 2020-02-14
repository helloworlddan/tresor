init:
    sudo apt-get install libgpgme-dev

build:
	go build -o tresor main.go

clean:
	rm -rf tresor
