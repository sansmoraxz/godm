.DEFAULT_GOAL := build
BIN_FILE=godm


build:
	go build -o bin/ ./cmd/godm/.
