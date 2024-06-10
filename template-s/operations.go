package main

// TODO: Implement the uploadTemplate, deleteTemplate, handleMessage, and consumeMessages functions
// TODO: Use common packages: minio, redis, and rabbitmq
type TemplateMessage struct {
	// TODO 1: Define the fields of the TemplateMessage struct
	// TODO 2: Add the json tags to the fields
}

func uploadTemplate(templateName, content string) error {
	// TODO 1: Connect to Minio and upload the Dockerfile content
	// TODO 2: Connect to Redis and save the template name and file path
	// TODO 3: Return an error if any of the operations fail

}

func deleteTemplate(templateName string) error {
	// TODO 1: Connect to Redis and get the file path of the template
	// TODO 2: Connect to Minio and delete the template file
	// TODO 3: Return an error if any of the operations fail
}

func handleMessage() {
	// TODO 1: Unmarshal the message body into a TemplateMessage struct
	// TODO 2: Call the appropriate function based on the operation
}

func consumeMessages() {
	// TODO 1: Connect to RabbitMQ and consume messages
	// TODO 2: Call the handleMessage function for each message
}
