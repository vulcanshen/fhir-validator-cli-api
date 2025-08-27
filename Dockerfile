FROM ubuntu AS builder
RUN apt update && apt install -y wget golang-go
RUN wget -P /opt/ https://github.com/hapifhir/org.hl7.fhir.core/releases/latest/download/validator_cli.jar
WORKDIR /go-build
COPY . .
RUN go build -o /opt/fhir-validator-cli-api

FROM ghcr.io/graalvm/jdk-community:21

WORKDIR /app
COPY --from=builder /opt/fhir-validator-cli-api ./fhir-validator-cli-api
COPY --from=builder /opt/validator_cli.jar ./validator_cli.jar

EXPOSE 8081

ENTRYPOINT ["/app/fhir-validator-cli-api"]
