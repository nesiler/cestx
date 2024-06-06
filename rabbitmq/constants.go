package rabbitmq

// Exchanges
const (
	ExchangeMachines   = "machines"
	ExchangeLogger     = "logger"
	ExchangeDynoxy     = "dynoxy"
	ExchangeTaskmaster = "taskmaster"
	ExchangeTemplate   = "template"
)

// Queues
const (
	QueueMachineCreate = "machine.create"
	QueueMachineDelete = "machine.delete"
	QueueMachineStart  = "machine.start"
	QueueMachineStop   = "machine.stop"
	QueueMachineUpdate = "machine.update"

	QueueTemplateCreate = "template.create"
	QueueTemplateDelete = "template.delete"
	QueueTemplateUpdate = "template.update"

	QueueLoggerMachine = "logger.machine"
	QueueLoggerDynoxy  = "logger.dynoxy"

	QueueDynoxyCreate = "dynoxy.create"
	QueueDynoxyDelete = "dynoxy.delete"

	QueueTaskmasterAnsible = "taskmaster.ansible"
	QueueTaskmasterSSH     = "taskmaster.ssh"
	QueueTaskmasterScript  = "taskmaster.script"
)
