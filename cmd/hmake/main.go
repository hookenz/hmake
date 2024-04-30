package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

type MakeArgs struct {
	debug   bool
	targets []string
}

// Makefile represents a parsed Makefile
type Makefile struct {
	Targets map[string]Target
}

// Target represents a target in the Makefile
type Target struct {
	Dependencies []string
	Commands     []string
}

func main() {
	// parse command line arguments

	args := ParseArgs()
	fmt.Println("Debug mode: ", args.debug)
	fmt.Println("Targets: ", args.targets)

	makefile := NewMakefile()
	err := makefile.Parse("Makefile")
	if err != nil {
		fmt.Println("Error parsing Makefile:", err)
		return
	}

	fmt.Println("Targets, dependencies, and commands:")
	for target, info := range makefile.Targets {
		fmt.Printf("Target: %s\n", target)
		fmt.Printf("  Dependencies: %s\n", strings.Join(info.Dependencies, ", "))
		fmt.Println("  Commands:")
		for _, cmd := range info.Commands {
			fmt.Printf("    %s\n", cmd)
		}
	}
}

func ParseArgs() MakeArgs {
	var args MakeArgs

	// Define flags
	debug := flag.Bool("d", false, "Enable debug mode")
	flag.Parse()

	// Targets are non-flag arguments
	targets := flag.Args()

	args.debug = *debug
	args.targets = targets

	return args
}

// NewMakefile initializes a new Makefile
func NewMakefile() *Makefile {
	return &Makefile{
		Targets: make(map[string]Target),
	}
}

// Parse parses a Makefile and populates the Makefile struct
func (mf *Makefile) Parse(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentTarget string
	var currentCommands []string
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// If it starts with a tab, it's a command
		if strings.HasPrefix(line, "\t") {
			currentCommands = append(currentCommands, strings.TrimSpace(line))
			continue
		}

		// Otherwise, it's a target
		if currentTarget != "" {
			// Save previous target and commands
			mf.Targets[currentTarget] = Target{
				Dependencies: mf.Targets[currentTarget].Dependencies,
				Commands:     currentCommands,
			}
			currentCommands = nil
		}

		parts := strings.Split(line, ":")
		currentTarget = strings.TrimSpace(parts[0])
		dependencies := []string{}

		// Extract dependencies if available
		if len(parts) > 1 {
			deps := strings.Split(parts[1], " ")
			for _, dep := range deps {
				dependencies = append(dependencies, strings.TrimSpace(dep))
			}
		}

		mf.Targets[currentTarget] = Target{
			Dependencies: dependencies,
			Commands:     nil,
		}
	}

	// Save commands of the last target
	if currentTarget != "" {
		mf.Targets[currentTarget] = Target{
			Dependencies: mf.Targets[currentTarget].Dependencies,
			Commands:     currentCommands,
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
