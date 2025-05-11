package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Launcher struct {
	executableDir string
	configPath    string
	serverPath    string
}

func NewLauncher() (*Launcher, error) {
	executableDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, fmt.Errorf("ошибка определения директории: %v", err)
	}

	return &Launcher{
		executableDir: executableDir,
	}, nil
}

func (l *Launcher) FindConfigPath() (string, error) {
	possibleConfigPaths := []string{
		filepath.Join(l.executableDir, "config"),
		filepath.Join(l.executableDir, "..", "config"),
		filepath.Join(filepath.Dir(l.executableDir), "config"),
		"/etc/notification-service/config",
		filepath.Join(os.Getenv("HOME"), ".notification-service/config"),
	}

	for _, path := range possibleConfigPaths {
		absPath, _ := filepath.Abs(path)
		if _, err := os.Stat(absPath); !os.IsNotExist(err) {
			return absPath, nil
		}
	}

	fmt.Println("ОШИБКА: Не найден путь к конфигурации! Проверьте наличие директории config.")
	fmt.Println("Проверенные пути:")
	for _, path := range possibleConfigPaths {
		absPath, _ := filepath.Abs(path)
		fmt.Printf("  - %s\n", absPath)
	}

	return "", fmt.Errorf("не найден путь к конфигурации")
}

func (l *Launcher) FindServerPath() (string, error) {
	serverPath := filepath.Join(l.executableDir, "bin", "server")
	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		serverPath = filepath.Join(l.executableDir, "server")
		if _, err := os.Stat(serverPath); os.IsNotExist(err) {
			fmt.Printf("ОШИБКА: Исполняемый файл сервера не найден!\n")
			fmt.Printf("Проверены пути:\n")
			fmt.Printf("  - %s\n", filepath.Join(l.executableDir, "bin", "server"))
			fmt.Printf("  - %s\n", filepath.Join(l.executableDir, "server"))
			return "", fmt.Errorf("исполняемый файл сервера не найден")
		}
	}

	return serverPath, nil
}

func (l *Launcher) RunServer() error {
	args := []string{}
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}

	cmd := exec.Command(l.serverPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	return cmd.Run()
}

func (l *Launcher) Initialize() error {
	var err error

	l.configPath, err = l.FindConfigPath()
	if err != nil {
		return err
	}
	fmt.Printf("Используется путь к конфигурации: %s\n", l.configPath)
	os.Setenv("CONFIG_PATH", l.configPath)

	l.serverPath, err = l.FindServerPath()
	if err != nil {
		return err
	}
	fmt.Printf("Используется исполняемый файл сервера: %s\n", l.serverPath)

	return nil
}

func main() {
	launcher, err := NewLauncher()
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	if err := launcher.Initialize(); err != nil {
		os.Exit(1)
	}

	if err := launcher.RunServer(); err != nil {
		fmt.Printf("Ошибка запуска приложения: %v\n", err)
		os.Exit(1)
	}
}
