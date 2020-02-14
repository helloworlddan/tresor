init:
	sudo apt-get install libgpgme11 libgpgme-dev libassuan0 libassuan-dev

build:
	go build -o tresor main.go

clean:
	rm -rf tresor
