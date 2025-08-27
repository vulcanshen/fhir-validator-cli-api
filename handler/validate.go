package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

type TestRequest struct {
	Payload json.RawMessage `json:"payload" swaggertype:"object"`
	Args    []string        `json:"args"`
}

func consoleOutputToPlainText(consoleOut string) string {
	return ansiRegex.ReplaceAllString(consoleOut, "")
}

// SyncValidation godoc
// @Summary Validate FHIR payload synchronously
// @Description Validates a FHIR payload using the validator_cli.jar and returns the result as plain text
// @Tags validation
// @Accept json
// @Produce plain
// @Param payload body TestRequest true "FHIR payload and arguments"
// @Success 200 {string} string "Validation output"
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal server error"
// @Router /validate [post]
func SyncValidation(c *gin.Context) {
	var reqBody TestRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Error parsing request body: %s", err.Error()))
		return
	}

	if len(reqBody.Payload) == 0 {
		c.String(http.StatusBadRequest, "Payload is required")
		return
	}

	tempPayloadFile, err := os.CreateTemp(".", "fhir_payload_*.json")

	_, err = io.Copy(tempPayloadFile, c.Request.Body)

	if err != nil {
		log.Printf("Error writing temp payload file: %v", err)
		c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to write payload to file: %s", err.Error()))
		return
	}

	_, err = tempPayloadFile.Write(reqBody.Payload)

	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to write payload to file: %s", err.Error()))
		return
	}

	_ = tempPayloadFile.Close()

	defer func(name string) {
		_ = os.Remove(name)
	}(tempPayloadFile.Name())

	log.Printf("Using temp file for fhir validator output: %s", tempPayloadFile.Name())

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	cmdArgs := []string{"-jar", "validator_cli.jar", tempPayloadFile.Name()}
	cmdArgs = append(cmdArgs, reqBody.Args...)

	cmdExecution := exec.CommandContext(ctx, "java", cmdArgs...)

	log.Printf("Starting executing validator_cli.jar: %s", strings.Join(cmdExecution.Args[1:], " "))
	output, err := cmdExecution.CombinedOutput()
	logOutput := string(output)
	if err != nil {
		log.Printf("validator_cli.jar command failed: %v", err)
		c.String(http.StatusInternalServerError, fmt.Sprintf("validator_cli.jar exited with status %s", err.Error()))
	} else {
		log.Println("validator_cli.jar command completed successfully.")
		//log.Printf("Command Output:\n%s", logOutput)

		c.String(http.StatusOK, consoleOutputToPlainText(logOutput))
	}
}

// SseValidation godoc
// @Summary Validate FHIR payload with SSE
// @Description Validates a FHIR payload using the validator_cli.jar and streams the output via Server-Sent Events (SSE)
// @Tags validation
// @Accept json
// @Produce text/event-stream
// @Param payload body TestRequest true "FHIR payload and arguments"
// @Success 200 {string} string "Validation output stream"
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal server error"
// @Router /validate/sse [post]
func SseValidation(c *gin.Context) {
	var reqBody TestRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Error parsing request body: %s", err.Error()))
		return
	}

	if len(reqBody.Payload) == 0 {
		c.String(http.StatusBadRequest, "Payload is required")
		return
	}

	tempPayloadFile, err := os.CreateTemp(".", "fhir_payload_*.json")
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to create temp file: %s", err.Error()))
		return
	}
	defer func() {
		_ = os.Remove(tempPayloadFile.Name())
	}()

	if _, err := tempPayloadFile.Write(reqBody.Payload); err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to write payload to file: %s", err.Error()))
		return
	}
	_ = tempPayloadFile.Close()

	log.Printf("Using temp file for fhir validator output: %s", tempPayloadFile.Name())

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	cmdArgs := []string{"-jar", "validator_cli.jar", tempPayloadFile.Name()}
	cmdArgs = append(cmdArgs, reqBody.Args...)

	cmdExecution := exec.CommandContext(ctx, "java", cmdArgs...)

	stdout, err := cmdExecution.StdoutPipe()
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to get stdout pipe: %s", err.Error()))
		return
	}
	stderr, err := cmdExecution.StderrPipe()
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to get stderr pipe: %s", err.Error()))
		return
	}

	lineChan := make(chan string)
	var wg sync.WaitGroup
	wg.Add(2)

	go streamOutput(stdout, ctx, lineChan, &wg)

	go streamOutput(stderr, ctx, lineChan, &wg)

	if err := cmdExecution.Start(); err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to start command: %s", err.Error()))
		return
	}

	go func() {
		wg.Wait()
		if err := cmdExecution.Wait(); err != nil {
			log.Printf("Command failed: %v", err)

			c.SSEvent("server_error", fmt.Sprintf("Command exited with error: %s", err.Error()))
		}
		close(lineChan)
	}()

	for {
		select {
		case msg, ok := <-lineChan:
			if !ok {

				log.Println("SSE stream finished.")
				return
			}

			c.SSEvent("console_out", msg)
			c.Writer.Flush()
		case <-c.Request.Context().Done():

			log.Println("Client disconnected, canceling command.")
			cancel()
			return
		}
	}
}

func streamOutput(pipe io.Reader, ctx context.Context, lineChan chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			lineChan <- consoleOutputToPlainText(scanner.Text())
		}
	}
}
