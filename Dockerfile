FROM ghcr.io/testssl/testssl.sh:3.2

COPY fhir-validator-cli-api /usr/local/bin/fhir-validator-cli-api
COPY validator_cli.jar /usr/local/bin/validator_cli.jar

EXPOSE 8081

ENTRYPOINT ["/usr/local/bin/fhir-validator-cli-api"]
