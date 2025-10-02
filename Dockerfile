FROM ghcr.io/graalvm/jdk-community:21

WORKDIR /app
COPY fhir-validator-cli-api ./fhir-validator-cli-api
#COPY validator_cli.jar ./validator_cli.jar
RUN curl -L -o validator_cli.jar https://github.com/hapifhir/org.hl7.fhir.core/releases/latest/download/validator_cli.jar

EXPOSE 8081

ENTRYPOINT ["/app/fhir-validator-cli-api"]
