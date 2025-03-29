package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

// runScript executes the given commands by piping them to the db executable
// and returns the output as a slice of strings (one per line)
func runScript(commands []string) ([]string, error) {
	// Create command to execute
	cmd := exec.Command("../bin/db")

	// Get pipes to stdin and stdout
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("error getting stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("error getting stdout pipe: %w", err)
	}

	// Start the command
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("error starting command: %w", err)
	}

	// Write commands to stdin
	go func() {
		defer stdin.Close()
		for _, command := range commands {
			io.WriteString(stdin, command+"\n")
		}
	}()

	// Read entire output
	var outputLines []string
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		outputLines = append(outputLines, scanner.Text())
	}

	// Wait for command to complete
	err = cmd.Wait()
	if err != nil {
		return outputLines, fmt.Errorf("command execution failed: %w", err)
	}

	return outputLines, nil
}

func TestDatabase(t *testing.T) {
	t.Run("inserts and retrieves a row", func(t *testing.T) {
		commands := []string{
			"insert 1 user1 person1@example.com",
			"select",
			".exit",
		}

		expected := []string{
			"db > Executed.",
			"db > (1, user1, person1@example.com)",
			"Executed.",
			"db > ",
		}

		result, err := runScript(commands)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		// Check if the result matches the expected output
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected: %v\nGot: %v", expected, result)
		}
	})

	t.Run("allows inserting strings that are the maximum length", func(t *testing.T) {
		longUsername := strings.Repeat("a", 32)
		longEmail := strings.Repeat("b", 255)

		commands := []string{
			fmt.Sprintf("insert 1 %s %s", longUsername, longEmail),
			"select",
			".exit",
		}

		expected := []string{
			"db > Executed.",
			fmt.Sprintf("db > (1, %s, %s)", longUsername, longEmail),
			"Executed.",
			"db > ",
		}

		result, err := runScript(commands)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		// Check if the result matches the expected output
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected: %v\nGot: %v", expected, result)
		}
	})

	t.Run("prints error message when strings are too long", func(t *testing.T) {
		longUsername := strings.Repeat("a", 33)
		longEmail := strings.Repeat("b", 256)

		commands := []string{
			fmt.Sprintf("insert 1 %s %s", longUsername, longEmail),
			"select",
			".exit",
		}

		expected := []string{
			"db > String is too long.",
			"db > Executed.",
			"db > ",
		}

		result, err := runScript(commands)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		// Check if the result matches the expected output
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected: %v\nGot: %v", expected, result)
		}
	})

	t.Run("prints an error message if id is negative", func(t *testing.T) {
		commands := []string{
			"insert -1 user1 person1@example.com",
			"select",
			".exit",
		}

		expected := []string{
			"db > ID must be positive.",
			"db > Executed.",
			"db > ",
		}

		result, err := runScript(commands)
		if err != nil {
			t.Fatalf("Failed to run script: %v", err)
		}

		// Check if the result matches the expected output
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected: %v\nGot: %v", expected, result)
		}
	})

	t.Run("prints error message when table is full", func(t *testing.T) {
		commands := []string{}
		for i := 0; i < 1401; i++ {
			command := fmt.Sprintf("insert %d user%d person%d@example.com", i, i, i)
			commands = append(commands, command)
		}

		expected := []string{
			"db > Error: Table full.",
		}

		result, _ := runScript(commands)

		// Check if the result matches the expected output
		actualResult := result[len(result)-2 : len(result)-1]
		if !reflect.DeepEqual(actualResult, expected) {
			t.Errorf("Expected: %v\nGot: %v", expected, actualResult)
		}
	})
}
