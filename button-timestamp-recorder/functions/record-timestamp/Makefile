build-RecordTimestampFunction:
	mkdir -p $(ARTIFACTS_DIR)
	# Ensure dependencies (AWS SigV4 signer) are available
	go mod download
	go get github.com/aws/aws-sdk-go-v2/aws/signer/v4@latest
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap db_insert.go
	chmod +x $(ARTIFACTS_DIR)/bootstrap

clean:
	rm -f bootstrap
