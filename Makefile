ifeq ($(OS),Windows_NT)
BINARY=hotreload.exe
SERVER_BINARY=testserver.exe
RUN_BINARY=.\\$(BINARY)
RUN_SERVER=.\\$(SERVER_BINARY)
SERVER_OUT=testserver\\$(SERVER_BINARY)
REMOVE=del /Q
else
BINARY=hotreload
SERVER_BINARY=testserver
RUN_BINARY=./$(BINARY)
RUN_SERVER=./$(SERVER_BINARY)
SERVER_OUT=testserver/$(SERVER_BINARY)
REMOVE=rm -f
endif

.PHONY: build demo clean

build:
	go build -o $(BINARY) ./cmd/hotreload

demo: build
	$(RUN_BINARY) --root ./testserver --build "go build -o $(SERVER_BINARY) ." --exec "$(RUN_SERVER)"

clean:
	-$(REMOVE) $(BINARY)
	-$(REMOVE) $(SERVER_OUT)
