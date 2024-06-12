package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"

	"github.com/Telmate/proxmox-api-go/cli"
	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	APIURL      string
	HTTPHeaders string
	User        string
	Password    string
	OTP         string
	NewCLI      bool
}

type LXCInfo struct {
	IP       string
	Hostname string
	VMID     int // Add a VMID for each LXC
}

var (
	systems = []LXCInfo{
		{"192.168.4.99", "MASTER", 9999},
		{"192.168.4.60", "REDIS", 9000},
		{"192.168.4.61", "POSTGRESQL", 9001},
		{"192.168.4.62", "RABBITMQ", 9002},
		{"192.168.4.70", "MINIO", 9010},
	}

	services = []LXCInfo{
		{"192.168.4.63", "REGISTRY", 9003},
		{"192.168.4.64", "LOGGER-S", 9004},
		{"192.168.4.65", "TASKMASTER-S", 9005},
		{"192.168.4.66", "DYNOXY-S", 9006},
		{"192.168.4.67", "TEMPLATE-S", 9007},
		{"192.168.4.68", "MACHINE-S", 9008},
		{"192.168.4.69", "API-GW", 9009},
	}

	hosts = []string{
		"redis",
		"postgresql",
		"rabbitmq",
		"minio",
	}
)

func Setup(client *proxmox.Client, fConfigFile *string) {
	var err error
	print("SETUP CESTX...\n")
	// homeDir, err := os.UserHomeDir()
	// if err != nil {
	// 	log.Fatal("Error getting home directory:", err)
	// }

	// cmd := exec.Command("rm", "-f", filepath.Join(homeDir, ".ssh", "known_hosts"))
	// cmd.Stdout = log.Writer()
	// cmd.Stderr = log.Writer()
	// err = cmd.Run()
	// failError(err)

	// // Create all containers
	// CreateAll(client, fConfigFile)
	// fmt.Print("All containers created successfully.\n")

	// // Start all containers
	// StartAll(client)
	// fmt.Print("All containers started successfully.\n")

	// // Run playbook: playbooks/setup-proxmox-api.yaml
	// fmt.Print("Setting up Proxmox API...\n")
	// err := runAnsiblePlaybook("playbooks/setup-proxmox-api.yaml", "all", map[string]string{"service": "master"})
	// failError(err)
	// fmt.Print("Proxmox API setup completed.\n")

	// // Run playbook for install docker on systems
	// fmt.Print("Installing Docker on systems...\n")
	// err = runAnsiblePlaybook("playbooks/docker.yaml", "systems", nil)
	// failError(err)

	// Run playbook for systems launch: playbooks/lanch.yaml
	fmt.Print("Launching systems...\n")
	for _, host := range hosts {
		err = runAnsiblePlaybook("playbooks/launch.yaml", host, map[string]string{"target": host})
		failError(err)
	}

	// // Run playbook for master host: playbooks/docker.yaml
	// fmt.Print("Installing Docker on master host...\n")
	// err = runAnsiblePlaybook("playbooks/portainer.yaml", "master", map[string]string{"service": "master"})
	// failError(err)

	// // Run playbook for master host: playbooks/ansible.yaml
	// fmt.Print("Installing Ansible on master host...\n")
	// err = runAnsiblePlaybook("playbooks/ansible.yaml", "master", map[string]string{"service": "master"})
	// failError(err)

	// Run playbook: playbooks/deploy.yaml
	fmt.Print("Deploying CESTX...\n")

	err = runAnsiblePlaybook("playbooks/deploy.yaml", "all", map[string]string{"service": "master"})
	failError(err)
	fmt.Print("CESTX deployed successfully.\n")
}

// runAnsiblePlaybook runs an Ansible playbook with optional extra variables
func runAnsiblePlaybook(playbookPath, host string, extraVars map[string]string) error {
	args := []string{"-l", host, playbookPath}
	for key, value := range extraVars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	// set inventory file
	args = append(args, "-i", "inventory/inventory.yaml")

	// common.Out("Running Ansible playbook: ansible-playbook %s", strings.Join(args, " "))
	cmd := exec.Command("ansible-playbook", args...)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	return cmd.Run()
}

func CreateAll(client *proxmox.Client, fConfigFile *string) {
	config, err := proxmox.NewConfigLxcFromJson(GetConfig(*fConfigFile))
	failError(err)

	sshKey := config.SSHPublicKeys
	if sshKey == "" {
		failError(proxmox.ErrorItemNotExists("sshkey", "yok yahu ?"))
	}

	for _, container := range systems {
		err := CreateLXC(client, &config, container.VMID, container.IP, container.Hostname, "16G")
		if err != nil {
			log.Println(err)
		}
		fmt.Printf("LXC %s created successfully.\n", container.Hostname)
	}

	for _, container := range services {
		// reduce memory and cores for services
		config.Memory = 512
		config.Cores = 1
		err := CreateLXC(client, &config, container.VMID, container.IP, container.Hostname, "8G")
		if err != nil {
			log.Println(err)
		}
		fmt.Printf("LXC %s created successfully.\n", container.Hostname)
	}

	fmt.Println("Containers created successfully.")
}

func StartAll(client *proxmox.Client) {
	for _, container := range systems {
		vmr := proxmox.NewVmRef(container.VMID)
		status, err := client.StartVm(vmr)
		failError(err)
		fmt.Printf("LXC %s started successfully: %s \n", container.Hostname, status)
	}

	for _, container := range services {
		vmr := proxmox.NewVmRef(container.VMID)
		status, err := client.StartVm(vmr)
		failError(err)
		fmt.Printf("LXC %s started successfully: %s \n", container.Hostname, status)
	}
}

func StopAll(client *proxmox.Client) {
	for _, container := range systems {
		vmr := proxmox.NewVmRef(container.VMID)
		status, err := client.StopVm(vmr)
		if err != nil {
			fmt.Printf("Error stopping LXC %s: %v\n", container.Hostname, err)
		}

		fmt.Printf("LXC %s stopped successfully: %s \n", container.Hostname, status)
	}

	for _, container := range services {
		vmr := proxmox.NewVmRef(container.VMID)
		status, err := client.StopVm(vmr)
		if err != nil {
			fmt.Printf("Error stopping LXC %s: %v\n", container.Hostname, err)
		}

		fmt.Printf("LXC %s stopped successfully: %s \n", container.Hostname, status)
	}
}

// Seperate CreateLXC function
func CreateLXC(client *proxmox.Client, config *proxmox.ConfigLxc, vmid int, ip string, hostname string, diskSize string) error {
	config.Hostname = hostname
	config.Networks[0]["ip"] = ip + "/22"

	// Get storage from parameter
	config.RootFs = proxmox.QemuDevice{
		"size":    diskSize,
		"storage": "local-lvm",
	}

	vmr := proxmox.NewVmRef(vmid)
	vmr.SetNode("nsp")

	err := config.CreateLxc(vmr, client)
	if err != nil {
		return err
	}

	return nil
}

func RemoveSetup(client *proxmox.Client) {
	print("Removing CESTX setup...\n")
	StopAll(client)
	for _, container := range services {
		vmr := proxmox.NewVmRef(container.VMID)

		result, err := client.DeleteVm(vmr)
		if err != nil {
			fmt.Printf("Error deleting LXC %s: %v\n", container.Hostname, err)
		} else {
			fmt.Printf("LXC %s deleted successfully.\n", container.Hostname)
		}

		fmt.Println(result)
	}

	for _, container := range systems {
		vmr := proxmox.NewVmRef(container.VMID)

		result, err := client.DeleteVm(vmr)
		if err != nil {
			fmt.Printf("Error deleting LXC %s: %v\n", container.Hostname, err)
		} else {
			fmt.Printf("LXC %s deleted successfully.\n", container.Hostname)
		}

		fmt.Println(result)
	}
}

func initializeProxmoxClient(config AppConfig, insecure bool, proxyURL string, taskTimeout int) (*proxmox.Client, error) {
	tlsconf := &tls.Config{InsecureSkipVerify: insecure}
	if !insecure {
		tlsconf = nil
	}

	client, err := proxmox.NewClient(config.APIURL, nil, config.HTTPHeaders, tlsconf, proxyURL, taskTimeout)
	if err != nil {
		return nil, err
	}

	if userRequiresAPIToken(config.User) {
		client.SetAPIToken(config.User, config.Password)
		_, err := client.GetVersion()
		if err != nil {
			return nil, err
		}
	} else {
		err = client.Login(config.User, config.Password, config.OTP)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

func failError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// GetConfig get config from file
func GetConfig(configFile string) (configSource []byte) {
	var err error
	if configFile != "" {
		configSource, err = os.ReadFile(configFile)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		configSource, err = io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
	}
	return
}

func loadAppConfig() AppConfig {
	newCLI := os.Getenv("NEW_CLI") == "true"

	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Failed to load .env file: %v", err)
	}

	return AppConfig{
		APIURL:      os.Getenv("PM_API_URL"),
		HTTPHeaders: os.Getenv("PM_HTTP_HEADERS"),
		User:        os.Getenv("PM_USER"),
		Password:    os.Getenv("PM_PASS"),
		OTP:         os.Getenv("PM_OTP"),
		NewCLI:      newCLI,
	}
}

var rxUserRequiresToken = regexp.MustCompile("[a-z0-9]+@[a-z0-9]+![a-z0-9]+")

func userRequiresAPIToken(userID string) bool {
	return rxUserRequiresToken.MatchString(userID)
}

func main() {
	config := loadAppConfig()

	if config.NewCLI {
		err := cli.Execute()
		if err != nil {
			failError(err)
		}
		os.Exit(0)
	}

	// Command-line flags
	insecure := flag.Bool("insecure", false, "TLS insecure mode")
	proxmox.Debug = flag.Bool("debug", false, "debug mode")
	fConfigFile := flag.String("file", "", "file to get the config from")
	taskTimeout := flag.Int("timeout", 300, "API task timeout in seconds")
	proxyURL := flag.String("proxy", "", "proxy URL to connect to")
	flag.Parse()

	// Initialize Proxmox client
	c, err := initializeProxmoxClient(config, *insecure, *proxyURL, *taskTimeout)
	if err != nil {
		log.Fatalf("Failed to initialize Proxmox client: %v", err)
	}

	// print config file
	if *fConfigFile != "" {
		config := GetConfig(*fConfigFile)
		cj, err := json.MarshalIndent(config, "", "  ")
		failError(err)
		log.Println(string(cj))
	}

	if len(flag.Args()) == 0 {
		fmt.Printf("Missing action, try start|stop vmid\n")
		os.Exit(0)
	}

	switch flag.Args()[0] {

	case "setup":
		Setup(c, fConfigFile)

	case "remove":
		print("Remove mode\n")
		RemoveSetup(c)

	default:
		fmt.Printf("unknown action, try start|stop vmid\n")
	}
}
