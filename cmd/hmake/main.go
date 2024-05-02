package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/dominikbraun/graph"
)

type MakeArgs struct {
	debug   bool
	targets []string
}

// Makefile represents a parsed Makefile
type Makefile struct {
	Targets   map[string]Target
	Variables map[string]string
}

// Target represents a target in the Makefile
type Target struct {
	Name         string
	Dependencies []string
	Commands     []string
}

// Run executes the commands of a target
func (t *Target) Run() {
	fmt.Println("running commands for target: ", t.Name)
	for _, command := range t.Commands {
		fmt.Println("    ", command)
	}
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

	targetHash := func(t Target) string {
		return t.Name
	}

	g := graph.New(targetHash, graph.Directed(), graph.Acyclic())
	for _, info := range makefile.Targets {
		if info.Name == ".PHONY" {
			continue
		}

		// fmt.Println("Adding vertex: ", info.Name)
		g.AddVertex(info)
	}

	for target, info := range makefile.Targets {
		for _, dep := range info.Dependencies {
			if target == ".PHONY" {
				continue
			}

			if err := g.AddEdge(target, dep); err != nil {
				panic(err)
			}
		}
	}

	for _, target := range args.targets {
		if _, ok := makefile.Targets[target]; !ok {
			fmt.Println("Target not found: ", target)
			os.Exit(1)
		}

		fmt.Println("Target: ", target)

		targets := []string{}

		graph.DFS(g, target, func(t string) bool {
			targets = append(targets, t)
			return false
		})

		// print reverse order
		for i := len(targets) - 1; i >= 0; i-- {
			// fmt.Println("  ", targets[i])
			t := makefile.Targets[targets[i]]
			t.Run()
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
		Targets:   make(map[string]Target),
		Variables: make(map[string]string),
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

		// Skip empty lines
		if line == "" || line[0] == '#' {
			continue
		}

		// If it starts with a tab, it's a command
		if strings.HasPrefix(line, "\t") {
			currentCommands = append(currentCommands, strings.TrimSpace(line))
			continue
		}

		// Check if line defines a variable
		if matches := regexp.MustCompile(`^(\w+)\s*=\s*(.*)$`).FindStringSubmatch(line); len(matches) == 3 {
			mf.Variables[matches[1]] = matches[2]
			continue
		}

		// Otherwise, it's a target
		if currentTarget != "" {
			// Save previous target and commands
			mf.Targets[currentTarget] = Target{
				Name:         currentTarget,
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
			// strip comments from the end of the dependancies list
			deps := parts[1]
			i := strings.Index(deps, "#")
			if i >= 0 {
				deps = deps[:i]
			}

			for _, dep := range strings.Split(deps, " ") {
				dep = strings.TrimSpace(dep)
				if dep != "" {
					dependencies = append(dependencies, strings.TrimSpace(dep))
				}
			}
		}

		mf.Targets[currentTarget] = Target{
			Name:         currentTarget,
			Dependencies: dependencies,
			Commands:     nil,
		}
	}

	// Save commands of the last target
	if currentTarget != "" {
		mf.Targets[currentTarget] = Target{
			Name:         currentTarget,
			Dependencies: mf.Targets[currentTarget].Dependencies,
			Commands:     currentCommands,
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
