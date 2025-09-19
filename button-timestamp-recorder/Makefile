.PHONY: build clean deploy validate local logs delete

build:
	sam build

clean:
	rm -rf functions/record-timestamp/bootstrap
	rm -rf .aws-sam

deploy:
	sam deploy --guided

validate:
	sam validate

local:
	sam local start-api

logs:
	sam logs -n RecordTimestampFunction --stack-name button-timestamp-recorder --tail

delete:
	sam delete --stack-name button-timestamp-recorder