package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// Определяем директорию с исполняемым файлом
	executableDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Printf("Ошибка определения директории: %v\n", err)
		os.Exit(1)
	}

	// Получаем аргументы командной строки, если они есть
	args := []string{}
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}

	// Запускаем основное приложение
	cmd := exec.Command(filepath.Join(executableDir, "bin", "server"), args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf("Ошибка запуска приложения: %v\n", err)
		os.Exit(1)
	}
}
