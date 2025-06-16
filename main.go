package main

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

func main() {
	fmt.Println()
	fmt.Println("####### SNITCH v0.01 #######")
	fmt.Println()
	printHostname()
	printOS()
	printCPU()
	printMemory()
	printUptime()
	printIPAddresses()
	checkDocker()
	checkUFW()
	readSQLiteDatabase()
	fmt.Println()
}

func printHostname() {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("Error getting hostname:", err)
		return
	}
	fmt.Println("Hostname:", hostname)
}

func printOS() {
	info, err := host.Info()
	if err != nil {
		fmt.Println("Error getting host info:", err)
		return
	}
	fmt.Printf("OS: %s %s (%s)\n", info.Platform, info.PlatformVersion, info.KernelVersion)
	fmt.Println("Architecture:", runtime.GOARCH)
}

func printCPU() {
	cpuInfo, err := cpu.Info()
	if err != nil {
		fmt.Println("Error getting CPU info:", err)
		return
	}
	if len(cpuInfo) > 0 {
		fmt.Printf("CPU: %s (%d cores)\n", cpuInfo[0].ModelName, runtime.NumCPU())
	}
}

func printMemory() {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println("Error getting memory info:", err)
		return
	}
	fmt.Printf("Memory: %.2f GB total, %.2f GB used (%.2f%%)\n",
		float64(vmStat.Total)/1e9,
		float64(vmStat.Used)/1e9,
		vmStat.UsedPercent)
}

func printUptime() {
	uptimeSec, err := host.Uptime()
	if err != nil {
		fmt.Println("Error getting uptime:", err)
		return
	}
	uptime := time.Duration(uptimeSec) * time.Second
	fmt.Println("Uptime:", uptime)
}

func printIPAddresses() {
	fmt.Println("IP Addresses:")
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Error getting interfaces:", err)
		return
	}
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}
			fmt.Printf(" - Interface: %s, IP: %s\n", iface.Name, ip.String())
		}
	}
}

func checkDocker() {
	fmt.Print("Docker Installed: ")
	_, err := exec.LookPath("docker")
	if err != nil {
		fmt.Println("No")
		return
	}
	fmt.Println("Yes")

	// Get Docker version
	cmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("  Unable to get Docker version")
		return
	}
	fmt.Printf("  Docker Version: %s\n", strings.TrimSpace(string(output)))

	// List containers
	fmt.Println("  Docker Containers:")
	listDockerContainers()
}

func listDockerContainers() {
	// Use docker ps to get container name, image, and status
	cmd := exec.Command("docker", "ps", "-a", "--format", "- {{.Names}} | {{.Status}} | {{.Image}}")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("  Error listing containers:", err)
		return
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fmt.Println("   ", line)
	}
}

func checkUFW() {
	if runtime.GOOS != "linux" {
		fmt.Println("UFW: Not applicable (non-Linux system)")
		return
	}

	fmt.Print("UFW Firewall Status: ")
	_, err := exec.LookPath("ufw")
	if err != nil {
		fmt.Println("Not installed")
		return
	}

	cmd := exec.Command("ufw", "status")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error checking UFW")
		return
	}

	status := strings.ToLower(string(output))
	if strings.Contains(status, "inactive") {
		fmt.Println("Inactive")
	} else if strings.Contains(status, "active") {
		fmt.Println("Active")
	} else {
		fmt.Println("Unknown status")
	}
}

func readSQLiteDatabase() {
	fmt.Println("NPM Records: ")

	startDir, err := os.Getwd()
	if err != nil {
		fmt.Println("  Error getting current directory:", err)
		return
	}

	fmt.Printf("  Searching for database.sqlite starting at: %s\n", startDir)

	err = filepath.Walk(startDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip unreadable directories/files (permissions)
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if info.Name() == "database.sqlite" {
			fmt.Printf("  Found database: %s\n", path)
			processSQLite(path)
			// Stop after first match
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		fmt.Println("  Error during file search:", err)
	}
}

func processSQLite(dbFile string) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		fmt.Println("  Error opening SQLite database:", err)
		return
	}
	defer db.Close()

	// Read proxy_host table
	fmt.Println("  Reading table: proxy_host")
	readProxyHostTable(db, "proxy_host")

	// Read redirection_host table
	fmt.Println("  Reading table: redirection_host")
	readRedirectionTable(db, "redirection_host")
}

func readProxyHostTable(db *sql.DB, tableName string) {
	rows, err := db.Query(fmt.Sprintf("SELECT is_deleted, domain_names, forward_host, forward_port FROM %s", tableName))
	if err != nil {
		fmt.Printf("    Error querying %s table: %v\n", tableName, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var isDeleted int
		var domainNames string
		var forwardHost string
		var forwardPort int

		err := rows.Scan(&isDeleted, &domainNames, &forwardHost, &forwardPort)
		if err != nil {
			fmt.Printf("    Error reading row from %s: %v\n", tableName, err)
			continue
		}

		status := ""
		if isDeleted == 1 {
			status = "[DELETED]"
		}

		fmt.Printf("    - %s -> %s:%d %s\n", domainNames, forwardHost, forwardPort, status)
	}
}

func readRedirectionTable(db *sql.DB, tableName string) {
	rows, err := db.Query(fmt.Sprintf("SELECT is_deleted, domain_names, forward_domain_name FROM %s", tableName))
	if err != nil {
		fmt.Printf("  Error querying %s table: %v\n", tableName, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var isDeleted int
		var domainNames string
		var forward_domain_name string

		err := rows.Scan(&isDeleted, &domainNames, &forward_domain_name)
		if err != nil {
			fmt.Printf("    Error reading row from %s: %v\n", tableName, err)
			continue
		}

		status := ""
		if isDeleted == 1 {
			status = "[DELETED]"
		}

		fmt.Printf("    - %s -> %s %s \n", domainNames, forward_domain_name, status)
	}
}
