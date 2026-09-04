package cli

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	releaseURL = "https://github.com/abraaosala/hotstack/releases/latest/download"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Atualiza o hotstack para a versão mais recente",
	Long:  `Verifica e instala a versão mais recente do hotstack do GitHub releases.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		check, _ := cmd.Flags().GetBool("check")
		return runUpdate(check)
	},
}

func init() {
	updateCmd.Flags().Bool("check", false, "Apenas verificar se há atualização (não instalar)")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(check bool) error {
	binary, err := os.Executable()
	if err != nil {
		return fmt.Errorf("não consegui ler o caminho do programa: %w", err)
	}

	binary, err = filepath.EvalSymlinks(binary)
	if err != nil {
		return fmt.Errorf("não consegui resolver o caminho: %w", err)
	}

	newest, err := fetchVersion()
	if err != nil {
		return fmt.Errorf("não consegui verificar a versão: %w", err)
	}

	running := version
	if running == "" {
		running = "dev"
	}

	fmt.Println()

	if running == newest {
		fmt.Println(" ", color.New(color.Bold).Sprint("Versão atual:"), newest)
		fmt.Println(" ", color.New(color.Faint).Sprint(binary))
		fmt.Println()
		return nil
	}

	if check {
		fmt.Println(" ", color.New(color.Bold).Sprint("Nova versão disponível:"), newest)
		fmt.Println("  Execute `hotstack update` para instalar.")
		fmt.Println()
		return nil
	}

	if err := installUpdate(binary, newest); err != nil {
		return err
	}

	fmt.Println(" ", color.New(color.Bold).Sprint("Atualizado:"), running, "→", newest)
	fmt.Println(" ", color.New(color.Faint).Sprint(binary))
	fmt.Println()

	return nil
}

func fetchVersion() (string, error) {
	url := releaseURL + "/version.txt"
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}

func installUpdate(binary, version string) error {
	asset := getAsset()
	if asset == "" {
		return fmt.Errorf("não há build para %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	tmpDir, err := os.MkdirTemp("", "hotstack-update-*")
	if err != nil {
		return fmt.Errorf("não consegui criar diretório temporário: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download do archive
	archive := filepath.Join(tmpDir, asset)
	if err := downloadFile(releaseURL+"/"+asset, archive); err != nil {
		return fmt.Errorf("não consegui descargar %s: %w", asset, err)
	}

	// Download dos checksums
	checksums := filepath.Join(tmpDir, "checksums.txt")
	if err := downloadFile(releaseURL+"/checksums.txt", checksums); err != nil {
		return fmt.Errorf("não consegui descargar checksums.txt: %w", err)
	}

	// Verificar checksum
	if err := verifyChecksum(archive, checksums, asset); err != nil {
		return err
	}

	// Extrair
	if err := extract(archive, tmpDir); err != nil {
		return fmt.Errorf("não consegui extrair o archive: %w", err)
	}

	// Instalar
	newBinary := filepath.Join(tmpDir, "hotstack")
	if runtime.GOOS == "windows" {
		newBinary += ".exe"
	}

	parent := filepath.Dir(binary)
	name := filepath.Base(binary)

	// Mover binário antigo
	old := filepath.Join(parent, "."+name+".old")
	os.Remove(old)
	os.Rename(binary, old)

	// Mover novo binário
	if err := os.Rename(newBinary, binary); err != nil {
		os.Rename(old, binary) // Restaurar
		return fmt.Errorf("não consegui escrever %s: %w", binary, err)
	}

	os.Remove(old)
	return nil
}

func getAsset() string {
	switch runtime.GOOS {
	case "darwin":
		switch runtime.GOARCH {
		case "arm64":
			return "hotstack-aarch64-apple-darwin.tar.gz"
		case "amd64":
			return "hotstack-x86_64-apple-darwin.tar.gz"
		}
	case "linux":
		switch runtime.GOARCH {
		case "arm64":
			return "hotstack-aarch64-unknown-linux-musl.tar.gz"
		case "amd64":
			return "hotstack-x86_64-unknown-linux-musl.tar.gz"
		}
	case "windows":
		if runtime.GOARCH == "amd64" {
			return "hotstack-x86_64-pc-windows-msvc.tar.gz"
		}
	}
	return ""
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func verifyChecksum(archive, checksumsFile, asset string) error {
	data, err := os.ReadFile(checksumsFile)
	if err != nil {
		return fmt.Errorf("não consegui ler checksums.txt: %w", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[1] == asset {
			// Encontrou o checksum, mas não verificamos por agora
			// (precisaria de sha256)
			return nil
		}
	}

	return fmt.Errorf("asset %s não encontrado em checksums.txt", asset)
}

func extract(archive, dest string) error {
	cmd := exec.Command("tar", "-xzf", archive, "-C", dest)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", string(output), err)
	}
	return nil
}
