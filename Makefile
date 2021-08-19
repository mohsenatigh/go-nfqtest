FLAG := -ldflags "-s -w"

release: FLAG = -ldflags "-s -w"
release: all

debug: FLAG = 
debug: all

all: simplefw.bin

simplefw.bin:
	go build $(FLAG) -o simplefw.bin *.go

clean:
	rm -f simplefw.bin
